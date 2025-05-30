import { arraysEqualMultiset } from "@/helpers/multiselect.ts";
import { SongForm } from "@/pages/SongPage/util/types.ts";

export const isNameValid = (name: string): boolean => {
  return !!name.trim() && !/^\s|\s$/.test(name);
};

export const isBpmValid = (bpm: string): boolean => {
  if (bpm.length === 0) return true; // Empty is considered valid (optional field)
  const bpmValue = parseInt(bpm, 10);
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
  // Compare primitive fields
  for (const key in formData) {
    if (Array.isArray(formData[key as keyof typeof formData])) {
      // Comparing arrays. We should add a specific comparison for arrays of objects providing func to compare those objects.
      return !arraysEqualMultiset(
        formData[key as keyof typeof formData] as Array<unknown>,
        initialFormData[key as keyof typeof initialFormData] as Array<unknown>,
        (a, b) => a === b,
      );
    }

    if (
      formData[key as keyof typeof formData] !==
      initialFormData[key as keyof typeof initialFormData]
    ) {
      return true;
    }
  }
  return false;
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
