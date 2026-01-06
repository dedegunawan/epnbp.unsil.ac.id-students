/// <reference types="vite/client" />

interface ImportMetaEnv {
  readonly VITE_TOKEN_KEY: string
  readonly VITE_SSO_LOGIN_URL: string
  readonly VITE_SSO_LOGOUT_URL: string
  readonly VITE_API_URL: string
  readonly VITE_BASE_URL: string
  readonly VITE_EPNBP_URL?: string
  readonly REDIRECT_ON_FAIL_PROFILE?: string
}

interface ImportMeta {
  readonly env: ImportMetaEnv
}





