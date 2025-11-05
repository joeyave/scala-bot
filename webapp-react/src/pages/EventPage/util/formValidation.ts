import { EventForm } from "@/pages/EventPage/util/types.ts";

export const isNameValid = (name: string): boolean => {
  return !!name.trim();
};

export const isDateValid = (date: string): boolean => {
  return !!date.trim();
};

export function isFormChanged(
  formData: EventForm,
  initialFormData: EventForm,
): boolean {
  // Normalize object by sorting any array values
  const normalize = (data: EventForm): Record<string, unknown> => {
    return Object.entries(data).reduce(
      (acc, [key, value]) => {
        if (Array.isArray(value)) {
          // Sort array to ignore order differences
          // eslint-disable-next-line @typescript-eslint/no-unsafe-assignment
          acc[key] = [...value].sort();
        } else {
          acc[key] = value;
        }
        return acc;
      },
      {} as Record<string, unknown>,
    );
  };

  const normalizedForm = normalize(formData);
  const normalizedInitial = normalize(initialFormData);

  // Compare via JSON stringification
  return JSON.stringify(normalizedForm) !== JSON.stringify(normalizedInitial);
}

export function isFormValid(formData: EventForm): boolean {
  return isNameValid(formData.name) && isDateValid(formData.date);
}
