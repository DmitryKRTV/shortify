import filter from "leo-profanity";

filter.loadDictionary("en");
filter.loadDictionary("ru");

function expandForScan(value: string): string {
  return value.toLowerCase().replace(/[./:?&=_-]+/g, " ");
}

export function hasProfanity(value: string): boolean {
  const text = value.trim();
  if (!text) return false;

  return filter.check(text) || filter.check(expandForScan(text));
}
