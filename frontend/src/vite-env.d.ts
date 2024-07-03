/// <reference types="vite/client" />

interface ImportMetaEnv {
    readonly VITE_FEATURES_SCENARIOS: boolean;
    // more env variables...
}

interface ImportMeta {
    readonly env: ImportMetaEnv;
}
