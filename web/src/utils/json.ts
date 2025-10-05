export const safeParseJSON = <T>(input: string): T | undefined => {
  try {
    return input ? (JSON.parse(input) as T) : undefined;
  } catch {
    return undefined;
  }
};
