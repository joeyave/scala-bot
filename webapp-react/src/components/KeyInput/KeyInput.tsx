import { Select, SelectProps } from "@telegram-apps/telegram-ui";
import React, { ChangeEvent, useEffect, useState } from "react";

const keyGroups = [
  {
    id: "major",
    label: "Major",
    options: [
      { value: "C", label: "C" },
      { value: "D", label: "D" },
      { value: "E", label: "E" },
      { value: "F", label: "F" },
      { value: "G", label: "G" },
      { value: "A", label: "A" },
      { value: "B", label: "B" },
    ],
  },
  {
    id: "majorSharp",
    label: "Major #",
    options: [
      // {value: "C#", label: "C#"},
      // {value: "D#", label: "D#"},
      { value: "F#", label: "F#" },
      // {value: "G#", label: "G#"},
      // {value: "A#", label: "A#"}
    ],
  },
  {
    id: "majorFlat",
    label: "Major b",
    options: [
      // {value: "Cb", label: "Cb"},
      { value: "Db", label: "Db" },
      { value: "Eb", label: "Eb" },
      { value: "Gb", label: "Gb" },
      { value: "Ab", label: "Ab" },
      { value: "Bb", label: "Bb" },
    ],
  },
  {
    id: "minor",
    label: "Minor",
    options: [
      { value: "Am", label: "Am" },
      { value: "Bm", label: "Bm" },
      { value: "Cm", label: "Cm" },
      { value: "Dm", label: "Dm" },
      { value: "Em", label: "Em" },
      { value: "Fm", label: "Fm" },
      { value: "Gm", label: "Gm" },
    ],
  },
  {
    id: "minorSharp",
    label: "Minor #",
    options: [
      // {value: "A#m", label: "A#m"},
      { value: "C#m", label: "C#m" },
      { value: "D#m", label: "D#m" },
      { value: "F#m", label: "F#m" },
      { value: "G#m", label: "G#m" },
    ],
  },
  {
    id: "minorFlat",
    label: "Minor b",
    options: [
      // {value: "Abm", label: "Abm"},
      { value: "Bbm", label: "Bbm" },
      // {value: "Dbm", label: "Dbm"},
      { value: "Ebm", label: "Ebm" },
      // {value: "Gbm", label: "Gbm"}
    ],
  },
];

// Create a flat list of all valid key values for validation
const allValidKeys = keyGroups.flatMap((group) =>
  group.options.map((option) => option.value),
);

// Export validation function
export function formatKey(input: string): string {
  return input;
}

interface KeyInputProps extends Omit<SelectProps, "onChange" | "children"> {
  value?: string;
  onChange?: (v: string) => void;
}

export const KeyInput: React.FC<KeyInputProps> = ({
  value = "",
  onChange,
  ...restProps
}) => {
  const [input, setInput] = useState(() => formatKey(value));

  // Update internal state when external value changes
  useEffect(() => {
    const formatted = formatKey(value);
    setInput(formatted);
  }, [value]);

  const handleInput = (e: ChangeEvent<HTMLSelectElement>) => {
    const formatted = formatKey(e.target.value);
    setInput(formatted);
    onChange?.(formatted);
  };

  return (
    <Select value={input} onChange={handleInput} {...restProps}>
      {!allValidKeys.includes(value) && (
        <option key={value} value={value}>
          {value}
        </option>
      )}

      {keyGroups.map((group) => (
        <optgroup key={group.id} label={group.label}>
          {group.options.map((option) => (
            <option key={option.value} value={option.value}>
              {option.label}
            </option>
          ))}
        </optgroup>
      ))}
    </Select>
  );
};
