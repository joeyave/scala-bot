import { allValidKeys, keyGroups } from "@/components/KeyInput/KeyInput.tsx";
import { Song } from "@/pages/EventPage/util/types.ts";
import { useSortable } from "@dnd-kit/sortable";
import { CSS } from "@dnd-kit/utilities";
import { TrashIcon } from "@heroicons/react/20/solid";
import { IconButton } from "@telegram-apps/telegram-ui";
import { hapticFeedback } from "@tma.js/sdk-react";
import React, { ChangeEvent } from "react";
import "./Setlist.css";

interface SetlistSongDisplayProps extends React.HTMLAttributes<HTMLDivElement> {
  song: Song;
  onRemove?: (song: Song) => void;
  onKeyChange?: (song: Song, newKey: string) => void;
  isOverlay?: boolean;
}

const SetlistSongDisplay = React.forwardRef<
  HTMLDivElement,
  SetlistSongDisplayProps
>(({ song, onRemove, onKeyChange, isOverlay, ...props }, ref) => {
  // Get the effective key (eventKey overrides original key)
  const effectiveKey = song.eventKey || song.key;
  const hasKeyOverride = song.eventKey && song.eventKey !== song.key;

  const handleKeyChange = (e: ChangeEvent<HTMLSelectElement>) => {
    e.stopPropagation();
    hapticFeedback.impactOccurred("light");
    onKeyChange?.(song, e.target.value);
  };

  return (
    <div
      className={`relative mb-1 flex items-center justify-between rounded-xl bg-[var(--tg-theme-section-bg-color)] px-4 py-3 ${isOverlay ? "z-10 cursor-grabbing shadow-lg" : "cursor-grab"} `}
      ref={ref}
      {...props}
    >
      <div className="mr-2 min-w-0 flex-1">
        <div className="mb-2 truncate font-[family-name:var(--tgui--font-family)] text-[length:var(--tgui--text--font_size)] leading-[var(--tgui--text--line_height)] font-[var(--tgui--font_weight--accent3)] text-[var(--tg-theme-text-color,#000)]">
          {song.name}
        </div>
        <div className="flex items-center truncate font-[family-name:var(--tgui--font-family)] text-[length:var(--tgui--subheadline2--font_size)] leading-[var(--tgui--subheadline2--line_height)] font-[var(--tgui--font_weight--accent3)] text-[var(--tg-theme-hint-color)]">
          {/* Key selector using native select */}
          <select
            value={effectiveKey}
            onChange={handleKeyChange}
            onClick={(e) => e.stopPropagation()}
            onPointerDown={(e) => e.stopPropagation()}
            onTouchStart={(e) => e.stopPropagation()}
            onMouseDown={(e) => e.stopPropagation()}
            className={`mr-2 w-16 cursor-pointer appearance-none rounded-lg bg-[length:12px] bg-[right_4px_center] bg-no-repeat py-0.5 pl-2 text-base font-semibold outline-none focus:outline-none ${
              hasKeyOverride
                ? "bg-[var(--tg-theme-secondary-bg-color)] bg-[url('data:image/svg+xml;charset=UTF-8,%3Csvg%20xmlns%3D%22http%3A%2F%2Fwww.w3.org%2F2000%2Fsvg%22%20viewBox%3D%220%200%2016%2016%22%20fill%3D%22white%22%3E%3Cpath%20fill-rule%3D%22evenodd%22%20d%3D%22M4.22%206.22a.75.75%200%200%201%201.06%200L8%208.94l2.72-2.72a.75.75%200%201%201%201.06%201.06l-3.25%203.25a.75.75%200%200%201-1.06%200L4.22%207.28a.75.75%200%200%201%200-1.06Z%22%20clip-rule%3D%22evenodd%22%2F%3E%3C%2Fsvg%3E')] text-[var(--tg-theme-text-color)]"
                : "bg-[var(--tg-theme-secondary-bg-color)] bg-[url('data:image/svg+xml;charset=UTF-8,%3Csvg%20xmlns%3D%22http%3A%2F%2Fwww.w3.org%2F2000%2Fsvg%22%20viewBox%3D%220%200%2016%2016%22%20fill%3D%22white%22%3E%3Cpath%20fill-rule%3D%22evenodd%22%20d%3D%22M4.22%206.22a.75.75%200%200%201%201.06%200L8%208.94l2.72-2.72a.75.75%200%201%201%201.06%201.06l-3.25%203.25a.75.75%200%200%201-1.06%200L4.22%207.28a.75.75%200%200%201%200-1.06Z%22%20clip-rule%3D%22evenodd%22%2F%3E%3C%2Fsvg%3E')] text-[var(--tg-theme-hint-color)]"
            }`}
          >
            {/* Show current value if not in valid keys */}
            {!allValidKeys.includes(effectiveKey) && (
              <option value={effectiveKey}>{effectiveKey}</option>
            )}

            {keyGroups
              .filter((group) => group.id !== "nashville")
              .map((group) => (
                <optgroup key={group.id} label={group.label}>
                  {group.options.map((option) => (
                    <option key={option.value} value={option.value}>
                      {option.label}
                    </option>
                  ))}
                </optgroup>
              ))}
          </select>
          <span>
            {song.bpm || "?"}, {song.time || "?"}
          </span>
        </div>
      </div>

      <IconButton
        onClick={(e) => {
          hapticFeedback.impactOccurred("light");
          e.stopPropagation();
          onRemove?.(song);
        }}
        mode="outline"
        size="s"
      >
        <TrashIcon className="h-5 w-5 text-[var(--tg-theme-destructive-text-color)]" />
      </IconButton>
    </div>
  );
});

export default SetlistSongDisplay;

SetlistSongDisplay.displayName = "SetlistSongDisplay";

interface SetlistSongProps {
  song: Song;
  onRemove: (id: Song) => void;
  onKeyChange?: (song: Song, newKey: string) => void;
}

export function SetlistSong({ song, onRemove, onKeyChange }: SetlistSongProps) {
  const {
    attributes,
    listeners,
    setNodeRef,
    transform,
    transition,
    isDragging,
  } = useSortable({ id: song.id });

  const style = {
    transform: CSS.Transform.toString(transform),
    transition,
    opacity: isDragging ? 0 : 1,
  };

  return (
    <SetlistSongDisplay
      ref={setNodeRef}
      style={style}
      song={song}
      onRemove={onRemove}
      onKeyChange={onKeyChange}
      {...attributes}
      {...listeners}
    />
  );
}
