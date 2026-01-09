import Search from "@/components/Setlist/Search.tsx";
import { Setlist } from "@/components/Setlist/Setlist.tsx";
import { Song } from "@/pages/EventPage/util/types.ts";
import { hapticFeedback } from "@tma.js/sdk-react";
import { Notify } from "notiflix";
import { useTranslation } from "react-i18next";
import { SectionHeader } from "@telegram-apps/telegram-ui/dist/components/Blocks/Section/components/SectionHeader/SectionHeader";

interface SetlistSectionProps {
  songs: Song[];
  onAddSong: (song: Song) => void;
  onRemove: (song: Song) => void;
  onReorder: (newSetlist: Song[]) => void;
  onKeyChange?: (song: Song, newKey: string) => void;
  driveFolderId: string;
  archiveFolderId?: string | null;
}

export function SetlistSection({
  songs,
  onAddSong,
  onRemove,
  onReorder,
  onKeyChange,
  driveFolderId,
  archiveFolderId,
}: SetlistSectionProps) {
  const { t } = useTranslation();

  const handleSelectSong = (song: Song) => {
    const alreadyExists = songs.some((s) => s.id === song.id);

    if (alreadyExists) {
      hapticFeedback.notificationOccurred("warning");
      Notify.failure(t("songExistsInSetlist"));
      return;
    }

    hapticFeedback.notificationOccurred("success");
    Notify.success(t("songAddedToSetlist"));
    onAddSong(song);
  };

  return (
    <div>
      <SectionHeader>{t("setlist")}</SectionHeader>
      <div className="mb-2">
        <Search
          driveFolderId={driveFolderId}
          archiveFolderId={archiveFolderId}
          onSelectSong={handleSelectSong}
        />
      </div>
      <Setlist items={songs} onRemove={onRemove} onReorder={onReorder} onKeyChange={onKeyChange} />
    </div>
  );
}
