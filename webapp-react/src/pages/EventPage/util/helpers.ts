export const getLocalDateTimeString = (
  date: Date,
  timezone?: string,
): string => {
  return date
    .toLocaleString("sv-SE", {
      timeZone: timezone, // Если undefined, используется системная зона
      year: "numeric",
      month: "2-digit",
      day: "2-digit",
      hour: "2-digit",
      minute: "2-digit",
      second: "2-digit",
    })
    .replace(" ", "T");
};
