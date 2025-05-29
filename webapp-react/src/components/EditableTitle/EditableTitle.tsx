// EditableTitle.tsx
import Icon28Check from "@/assets/icons/check_28.svg";
import { AutosizeTextarea } from "@/components/AutosizeTextarea/AutosizeTextarea.tsx";
import { IconButton, Title } from "@telegram-apps/telegram-ui";
import { Icon28Edit } from "@telegram-apps/telegram-ui/dist/icons/28/edit";
import React, { useEffect, useState } from "react";

interface EditableTitleProps {
  value: string;
  onChange: (value: string) => void;
  titleLevel?: "1" | "2" | "3";
  titleWeight?: "1" | "2" | "3";
  maxHeight?: string;
  status?: "error" | undefined;
}

// regex to catch:
//   • Windows-forbidden:  \ / : * ? " < > |
//   • macOS-forbidden:    :
//   • Linux-forbidden:    null (\x00) and slash (/)
// (we include control-chars \x00–\x1F, too, for extra safety)

// const INVALID_FILENAME_CHARS = /[\x00-\x1F\\/:*?"<>|]/g;

// eslint-disable-next-line no-control-regex
const INVALID_FILENAME_CHARS = /[\x00-\x1F\\/:*<>]/g;

export function formatTitle(name: string): string {
  const stripped = name.replace(INVALID_FILENAME_CHARS, "");
  // 2) (Optionally) trim length so you don't hit OS path limits
  return stripped.slice(0, 255).trimStart();
}

export const EditableTitle: React.FC<EditableTitleProps> = ({
  value,
  onChange,
  titleLevel = "1",
  titleWeight = "2",
  maxHeight = "100px",
  status,
}) => {
  const [isEditing, setIsEditing] = useState<boolean>(false);
  const [input, setInput] = useState(() => formatTitle(value));

  // Update internal state when external value changes
  useEffect(() => {
    const formatted = formatTitle(value);
    setInput(formatted);
  }, [value]);

  const handleInput = (e: React.ChangeEvent<HTMLInputElement>) => {
    const formatted = formatTitle(e.target.value);
    setInput(formatted);
    onChange?.(formatted);
  };

  return (
    <div className="flex content-between items-center">
      <div className="mr-2 flex-1">
        {isEditing ? (
          <AutosizeTextarea
            value={input}
            onChange={handleInput}
            onBlur={() => {
              if (status != "error") {
                setIsEditing(false);
              }
            }}
            onKeyDown={(e) => {
              if (e.key === "Enter" && !e.shiftKey) {
                e.preventDefault();
                if (status != "error") {
                  setIsEditing(false);
                }
              }
            }}
            autoFocus
            maxHeight={maxHeight}
            status={status}
          />
        ) : (
          <Title level={titleLevel} weight={titleWeight}>
            {input}
          </Title>
        )}
      </div>
      <IconButton
        mode="plain"
        size="s"
        onClick={() => {
          if (status != "error") {
            setIsEditing(!isEditing);
          }
        }}
      >
        {isEditing ? <Icon28Check /> : <Icon28Edit />}
      </IconButton>
    </div>
  );
};
