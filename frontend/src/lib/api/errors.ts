export interface ApiDiagnostic {
  kind: "config" | "auth" | "network" | "http" | "unknown";
  title: string;
  detail: string;
  retryable: boolean;
  status?: number;
  code?: string;
  digest?: string;
}

export class ApiConfigError extends Error {
  diagnostic: ApiDiagnostic;

  constructor(detail: string, code = "API_CONFIG_ERROR") {
    super(detail);
    this.name = "ApiConfigError";
    this.diagnostic = {
      kind: "config",
      title: "Configuration issue",
      detail,
      retryable: false,
      code,
    };
  }
}

export class ApiRequestError extends Error {
  diagnostic: ApiDiagnostic;

  constructor(diagnostic: ApiDiagnostic) {
    super(diagnostic.detail);
    this.name = "ApiRequestError";
    this.diagnostic = diagnostic;
  }
}

export function toApiDiagnostic(error: unknown): ApiDiagnostic {
  if (error instanceof ApiConfigError || error instanceof ApiRequestError) {
    return error.diagnostic;
  }

  if (error instanceof Error) {
    return {
      kind: "unknown",
      title: "Unexpected failure",
      detail: error.message || "An unexpected failure interrupted rendering.",
      retryable: true,
      digest: "digest" in error && typeof error.digest === "string"
        ? error.digest
        : undefined,
    };
  }

  return {
    kind: "unknown",
    title: "Unexpected failure",
    detail: "An unknown error interrupted rendering.",
    retryable: true,
  };
}
