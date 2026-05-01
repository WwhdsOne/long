package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"long/internal/vote"
)

const (
	defaultTasksCollection    = "task_definitions"
	defaultArchivesCollection = "task_cycle_archives"
)

type cliOptions struct {
	command            string
	mongoURI           string
	database           string
	tasksCollection    string
	archivesCollection string
	connectTimeout     time.Duration
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	opts, err := parseArgs(os.Args[1:])
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), opts.connectTimeout)
	defer cancel()
	return runWithOptions(ctx, opts)
}

func parseArgs(args []string) (cliOptions, error) {
	if len(args) == 0 {
		return cliOptions{}, errors.New("用法: go -C backend run ./cmd/migratetaskmodel <plan|migrate|verify> --mongo-uri <uri> --db <database> [--tasks-collection task_definitions] [--archives-collection task_cycle_archives]")
	}

	opts := cliOptions{
		command:            strings.ToLower(strings.TrimSpace(args[0])),
		tasksCollection:    defaultTasksCollection,
		archivesCollection: defaultArchivesCollection,
		connectTimeout:     15 * time.Second,
	}
	flagSet := flag.NewFlagSet("migratetaskmodel", flag.ContinueOnError)
	flagSet.SetOutput(os.Stderr)
	flagSet.StringVar(&opts.mongoURI, "mongo-uri", "", "MongoDB URI")
	flagSet.StringVar(&opts.database, "db", "", "MongoDB database name")
	flagSet.StringVar(&opts.tasksCollection, "tasks-collection", defaultTasksCollection, "任务定义集合名")
	flagSet.StringVar(&opts.archivesCollection, "archives-collection", defaultArchivesCollection, "任务归档集合名")
	flagSet.DurationVar(&opts.connectTimeout, "connect-timeout", 15*time.Second, "MongoDB 连接超时")
	if err := flagSet.Parse(args[1:]); err != nil {
		return cliOptions{}, err
	}
	if opts.command != "plan" && opts.command != "migrate" && opts.command != "verify" {
		return cliOptions{}, fmt.Errorf("未知命令 %q", opts.command)
	}
	if strings.TrimSpace(opts.mongoURI) == "" {
		return cliOptions{}, errors.New("必须显式提供 --mongo-uri")
	}
	if strings.TrimSpace(opts.database) == "" {
		return cliOptions{}, errors.New("必须显式提供 --db")
	}
	if strings.TrimSpace(opts.tasksCollection) == "" {
		return cliOptions{}, errors.New("必须显式提供 --tasks-collection")
	}
	if strings.TrimSpace(opts.archivesCollection) == "" {
		return cliOptions{}, errors.New("必须显式提供 --archives-collection")
	}
	return opts, nil
}

func runWithOptions(ctx context.Context, opts cliOptions) error {
	client, err := mongo.Connect(ctx, options.Client().
		ApplyURI(opts.mongoURI).
		SetConnectTimeout(opts.connectTimeout))
	if err != nil {
		return fmt.Errorf("connect mongo: %w", err)
	}
	defer client.Disconnect(context.Background())

	db := client.Database(opts.database)
	tasks := db.Collection(opts.tasksCollection)
	archives := db.Collection(opts.archivesCollection)

	switch opts.command {
	case "plan":
		return runPlan(ctx, tasks, archives)
	case "migrate":
		return runMigrate(ctx, tasks, archives)
	case "verify":
		return runVerify(ctx, tasks, archives)
	default:
		return fmt.Errorf("未知命令 %q", opts.command)
	}
}

func runPlan(ctx context.Context, tasks *mongo.Collection, archives *mongo.Collection) error {
	taskCount, err := countPendingDocuments(ctx, tasks)
	if err != nil {
		return err
	}
	archiveCount, err := countPendingDocuments(ctx, archives)
	if err != nil {
		return err
	}
	fmt.Printf("待迁移任务定义: %d\n", taskCount)
	fmt.Printf("待迁移任务归档: %d\n", archiveCount)
	return nil
}

func runMigrate(ctx context.Context, tasks *mongo.Collection, archives *mongo.Collection) error {
	taskCount, err := migrateCollection(ctx, tasks, normalizeTaskDefinitionDocument)
	if err != nil {
		return err
	}
	archiveCount, err := migrateCollection(ctx, archives, normalizeTaskArchiveDocument)
	if err != nil {
		return err
	}
	fmt.Printf("已迁移任务定义: %d\n", taskCount)
	fmt.Printf("已迁移任务归档: %d\n", archiveCount)
	return nil
}

func runVerify(ctx context.Context, tasks *mongo.Collection, archives *mongo.Collection) error {
	taskCount, err := countPendingDocuments(ctx, tasks)
	if err != nil {
		return err
	}
	archiveCount, err := countPendingDocuments(ctx, archives)
	if err != nil {
		return err
	}
	if taskCount > 0 || archiveCount > 0 {
		return fmt.Errorf("仍有文档未补齐字段: tasks=%d archives=%d", taskCount, archiveCount)
	}
	fmt.Println("校验通过：所有任务定义和任务归档都已补齐 event_kind/window_kind")
	return nil
}

func countPendingDocuments(ctx context.Context, collection *mongo.Collection) (int64, error) {
	count, err := collection.CountDocuments(ctx, pendingMigrationFilter())
	if err != nil {
		return 0, fmt.Errorf("count %s: %w", collection.Name(), err)
	}
	return count, nil
}

func migrateCollection(ctx context.Context, collection *mongo.Collection, normalizer func(bson.M) (bson.M, bool)) (int64, error) {
	cursor, err := collection.Find(ctx, pendingMigrationFilter())
	if err != nil {
		return 0, fmt.Errorf("find %s: %w", collection.Name(), err)
	}
	defer cursor.Close(ctx)

	var migrated int64
	for cursor.Next(ctx) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			return migrated, fmt.Errorf("decode %s: %w", collection.Name(), err)
		}
		update, changed := normalizer(doc)
		if !changed {
			continue
		}
		if _, err := collection.UpdateByID(ctx, doc["_id"], bson.M{"$set": update}); err != nil {
			return migrated, fmt.Errorf("update %s %v: %w", collection.Name(), doc["_id"], err)
		}
		migrated++
	}
	if err := cursor.Err(); err != nil {
		return migrated, fmt.Errorf("iterate %s: %w", collection.Name(), err)
	}
	return migrated, nil
}

func pendingMigrationFilter() bson.M {
	return bson.M{
		"$or": []bson.M{
			{"event_kind": bson.M{"$exists": false}},
			{"event_kind": ""},
			{"window_kind": bson.M{"$exists": false}},
			{"window_kind": ""},
		},
	}
}

func normalizeTaskDefinitionDocument(doc bson.M) (bson.M, bool) {
	item := vote.NormalizeTaskDefinitionModel(vote.TaskDefinition{
		TaskID:        stringValue(doc["task_id"]),
		TaskType:      vote.TaskType(stringValue(doc["task_type"])),
		EventKind:     vote.TaskEventKind(stringValue(doc["event_kind"])),
		WindowKind:    vote.TaskWindowKind(stringValue(doc["window_kind"])),
		ConditionKind: vote.TaskConditionKind(stringValue(doc["condition_kind"])),
	})
	update := bson.M{
		"task_type":      string(item.TaskType),
		"event_kind":     string(item.EventKind),
		"window_kind":    string(item.WindowKind),
		"condition_kind": string(item.ConditionKind),
	}
	return update, requiresMigration(doc, update)
}

func normalizeTaskArchiveDocument(doc bson.M) (bson.M, bool) {
	item := vote.NormalizeTaskArchiveModel(vote.TaskCycleArchive{
		TaskID:        stringValue(doc["task_id"]),
		TaskType:      vote.TaskType(stringValue(doc["task_type"])),
		EventKind:     vote.TaskEventKind(stringValue(doc["event_kind"])),
		WindowKind:    vote.TaskWindowKind(stringValue(doc["window_kind"])),
		ConditionKind: vote.TaskConditionKind(stringValue(doc["condition_kind"])),
	})
	update := bson.M{
		"task_type":      string(item.TaskType),
		"event_kind":     string(item.EventKind),
		"window_kind":    string(item.WindowKind),
		"condition_kind": string(item.ConditionKind),
	}
	return update, requiresMigration(doc, update)
}

func requiresMigration(doc bson.M, update bson.M) bool {
	for key, value := range update {
		if stringValue(doc[key]) != stringValue(value) {
			return true
		}
	}
	return false
}

func stringValue(value any) string {
	if value == nil {
		return ""
	}
	switch typed := value.(type) {
	case string:
		return strings.TrimSpace(typed)
	case fmt.Stringer:
		return strings.TrimSpace(typed.String())
	default:
		return strings.TrimSpace(fmt.Sprintf("%v", value))
	}
}
