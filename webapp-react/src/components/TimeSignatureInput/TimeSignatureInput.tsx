// TimeSignatureInput.tsx
import { Input, InputProps } from "@telegram-apps/telegram-ui";
import React, { useEffect, useState } from "react";

const COMMON_SIGNATURES = ["4/4", "3/4", "6/8", "2/4"];

// Export validation function
export function formatTimeSignature(input: string): string {
  if (!input) return "";

  // Formatting and validation logic moved from handleInput
  let formatted = "";
  let slashAdded = false;
  for (let i = 0; i < input.length && formatted.length < 5; i++) {
    const c = input[i];
    if (/\d/.test(c)) {
      formatted += c;
    } else if (!slashAdded && formatted.length > 0 && formatted.length < 4) {
      formatted += "/";
      slashAdded = true;
    }
  }
  const parts = formatted.split("/");
  if (parts.length > 2) {
    formatted = parts[0] + "/" + parts.slice(1).join("").slice(0, 2);
  }

  return formatted;
}

interface TimeSignatureInputProps extends Omit<InputProps, "onChange"> {
  value?: string;
  onChange?: (v: string) => void;
  placeholder?: string;
}

export const TimeSignatureInput: React.FC<TimeSignatureInputProps> = ({
  value = "",
  onChange,
  placeholder = "Time",
  ...restProps
}) => {
  const [input, setInput] = useState(() => formatTimeSignature(value));

  // Update internal state when external value changes
  useEffect(() => {
    const formatted = formatTimeSignature(value);
    setInput(formatted);
  }, [value]);

  const handleInput = (e: React.ChangeEvent<HTMLInputElement>) => {
    const formatted = formatTimeSignature(e.target.value);
    setInput(formatted);
    onChange?.(formatted);
  };

  return (
    <>
      <Input
        value={input}
        onChange={handleInput}
        placeholder={placeholder}
        maxLength={5}
        inputMode="text"
        list="signature-hints"
        autoComplete="off"
        {...restProps}
      />
      <datalist id="signature-hints">
        {COMMON_SIGNATURES.map((sig) => (
          <option key={sig} value={sig} />
        ))}
      </datalist>
    </>
  );
};
