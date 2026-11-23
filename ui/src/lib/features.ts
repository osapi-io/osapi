const env = import.meta.env;

export const features = {
  keyboard: env.OSAPI_FEATURE_KEYBOARD !== "false",
} as const;
