export interface SSEEventPayload {
  event: string;
  data: string;
}

export const parseSSEEvent = (chunk: string): SSEEventPayload | null => {
  let eventName = 'message';
  const dataLines: string[] = [];

  const lines = chunk.split(/\r?\n/);
  for (const rawLine of lines) {
    const line = rawLine.trimEnd();
    if (!line) {
      continue;
    }
    if (line.startsWith(':')) {
      continue;
    }
    if (line.startsWith('event:')) {
      eventName = line.slice(6).trim();
      continue;
    }
    if (line.startsWith('data:')) {
      const value = line.slice(5);
      dataLines.push(value.startsWith(' ') ? value.slice(1) : value);
    }
  }

  return {
    event: eventName || 'message',
    data: dataLines.join('\n'),
  };
};
