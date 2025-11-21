export const getLocalDateTimeString = (
  date: Date,
  noTime: boolean = false,
  timezone?: string,
): string => {
  if (noTime) {
    date.setHours(0, 0, 0, 0);
  }
  return date
    .toLocaleString("sv-SE", {
      timeZone: timezone, // Если undefined, используется системная зона
      year: "numeric",
      month: "2-digit",
      day: "2-digit",
      hour: "2-digit",
      minute: "2-digit",
    })
    .replace(" ", "T");
};
