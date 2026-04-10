const env = import.meta.env;

export const features = {
  stacks: env.OSAPI_FEATURE_STACKS === "true",
  keyboard: env.OSAPI_FEATURE_KEYBOARD !== "false",
} as const;
