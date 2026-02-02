import { SongForm } from "@/pages/SongPage/util/types.ts";

export const isNameValid = (name: string): boolean => {
  return !!name.trim() && !/^\s|\s$/.test(name);
};

export const isBpmValid = (bpm: string): boolean => {
  if (bpm.length === 0) return true; // Empty is considered valid (optional field)

  // Pattern: 1-3 digits, optionally followed by exactly ".5"
  const bpmPattern = /^\d{1,3}(\.5)?$/;
  if (!bpmPattern.test(bpm)) return false;

  const bpmValue = parseFloat(bpm);
  return !isNaN(bpmValue) && bpmValue >= 20 && bpmValue <= 300;
};

export const isTimeSignatureValid = (time: string): boolean => {
  if (time.length === 0) return true; // Empty is considered valid (optional field)
  return /^\d{1,2}\/\d{1,2}$/.test(time);
};

export function isFormChanged(
  formData: SongForm,
  initialFormData: SongForm,
): boolean {
  // Normalize object by sorting any array values
  const normalize = (data: SongForm): Record<string, unknown> => {
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

export function isFormValid(
  formData: SongForm,
  transpositionError: boolean,
): boolean {
  return (
    isNameValid(formData.name) &&
    isBpmValid(formData.bpm) &&
    isTimeSignatureValid(formData.time) &&
    !transpositionError
  );
}
