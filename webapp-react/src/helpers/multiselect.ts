export function arraysEqualMultiset<T>(
  a: T[],
  b: T[],
  isEqual: (a: T, b: T) => boolean,
): boolean {
  if (a.length !== b.length) return false;

  // Copy so we can mutate
  const bCopy = [...b];

  for (const itemA of a) {
    // Find index in bCopy where isEqual is true
    const idx = bCopy.findIndex((itemB) => isEqual(itemA, itemB));
    if (idx === -1) return false; // Not found
    bCopy.splice(idx, 1); // Remove matched item (so counts are correct)
  }

  // If all matched, bCopy should be empty
  return bCopy.length === 0;
}
