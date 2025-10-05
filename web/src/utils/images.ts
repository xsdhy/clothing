export const sanitizeImages = (images?: string[]): string[] =>
  (images ?? [])
    .map((image) => (typeof image === 'string' ? image.trim() : ''))
    .filter((image): image is string => image.length > 0);

export const appendSanitizedImages = (target: Set<string>, images?: string[]): void => {
  sanitizeImages(images).forEach((item) => {
    if (!target.has(item)) {
      target.add(item);
    }
  });
};
