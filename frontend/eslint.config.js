import js from '@eslint/js'
import globals from 'globals'
import vue from 'eslint-plugin-vue'
import unusedImports from 'eslint-plugin-unused-imports'

const sharedGlobals = {
    ...globals.browser,
    ...globals.node,
    ...globals.es2024,
}

const testGlobals = {
    ...sharedGlobals,
    afterAll: 'readonly',
    afterEach: 'readonly',
    beforeAll: 'readonly',
    beforeEach: 'readonly',
    describe: 'readonly',
    expect: 'readonly',
    it: 'readonly',
    test: 'readonly',
    vi: 'readonly',
}

export default [
    {
        ignores: ['backend/public/**'],
    },
    js.configs.recommended,
    ...vue.configs['flat/essential'],
    {
        linterOptions: {
            reportUnusedDisableDirectives: 'off',
        },
    },
    {
        files: ['**/*.{js,mjs,cjs,vue}'],
        languageOptions: {
            ecmaVersion: 'latest',
            sourceType: 'module',
            globals: sharedGlobals,
        },
        plugins: {
            'unused-imports': unusedImports,
        },
        rules: {
            'vue/no-mutating-props': 'off',
            'no-unused-vars': 'off',
            'unused-imports/no-unused-imports': 'error',
        },
    },
    {
        files: ['**/*.test.js', '**/*.test.mjs', '**/*.test.cjs', '**/*.spec.js', '**/*.spec.mjs', '**/*.spec.cjs'],
        languageOptions: {
            globals: testGlobals,
        },
    },
]
