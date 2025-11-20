const VIDEO_EXTENSIONS = [".mp4", ".webm", ".mov", ".mkv", ".avi", ".m4v"];

export const isVideoUrl = (src?: string): boolean => {
  const value = src?.trim();
  if (!value) {
    return false;
  }
  const lower = value.toLowerCase();
  if (lower.startsWith("data:video/")) {
    return true;
  }
  for (const ext of VIDEO_EXTENSIONS) {
    if (lower.endsWith(ext)) {
      return true;
    }
  }
  try {
    const url = new URL(lower);
    const pathname = url.pathname.toLowerCase();
    return VIDEO_EXTENSIONS.some((ext) => pathname.endsWith(ext));
  } catch {
    return false;
  }
};

export const buildDownloadName = (
  src: string,
  fallbackBase: string,
  defaultExt: string,
): string => {
  const trimmed = src.trim();
  let ext = "";
  try {
    const url = new URL(trimmed);
    const path = url.pathname;
    const last = path.split("/").pop() || "";
    const dot = last.lastIndexOf(".");
    if (dot !== -1 && dot < last.length - 1) {
      ext = last.slice(dot);
    }
  } catch {
    const dot = trimmed.lastIndexOf(".");
    if (dot !== -1 && dot < trimmed.length - 1) {
      ext = trimmed.slice(dot);
    }
  }
  const safeExt = ext || defaultExt;
  const base =
    fallbackBase +
    (safeExt.startsWith(".") ? safeExt : safeExt ? `.${safeExt}` : "");
  return base;
};
