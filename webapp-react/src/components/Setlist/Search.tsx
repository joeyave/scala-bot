import * as React from "react";
import { Autocomplete } from "@base-ui-components/react/autocomplete";
import { searchDriveFiles } from "@/api/webapp/driveFiles.ts";
import { DriveFile } from "@/api/webapp/typesResp.ts";
import { MagnifyingGlassIcon } from "@heroicons/react/16/solid";
import { IconButton } from "@telegram-apps/telegram-ui";
import { getSongByDriveFileId } from "@/api/webapp/songs.ts";
import { hapticFeedback } from "@tma.js/sdk-react";
import { Song } from "@/pages/EventPage/util/types.ts";
import { useTranslation } from "react-i18next";
import { Notify } from "notiflix";

interface SearchProps {
  driveFolderId: string;
  archiveFolderId?: string | null;
  onSelectSong?: (song: Song) => void;
}

export default function Search({
  onSelectSong,
  driveFolderId,
  archiveFolderId,
}: SearchProps) {
  const { t } = useTranslation();

  const [searchValue, setSearchValue] = React.useState("");
  const [isLoading, setIsLoading] = React.useState(false);
  const [searchResults, setSearchResults] = React.useState<DriveFile[]>([]);
  const [isAddingSong, setIsAddingSong] = React.useState(false);
  const [error, setError] = React.useState<string | null>(null);

  const wrapperRef = React.useRef<HTMLDivElement>(null);

  React.useEffect(() => {
    if (!searchValue && !isAddingSong) {
      setSearchResults([]);
      setIsLoading(false);
      return undefined;
    }

    setIsLoading(true);
    setError(null);

    let ignore = false;

    async function fetchDriveFiles() {
      try {
        const results = await searchDriveFiles(
          searchValue,
          driveFolderId,
          archiveFolderId,
        );

        if (!ignore) {
          hapticFeedback.impactOccurred("light");
          setSearchResults(results);
        }

        // eslint-disable-next-line @typescript-eslint/no-unused-vars
      } catch (err) {
        if (!ignore) {
          // todo: localize.
          setError(t("errorFetchingSongs"));
          setSearchResults([]);
        }
      } finally {
        if (!ignore) {
          setIsLoading(false);
        }
      }
    }

    const timeoutId = setTimeout(fetchDriveFiles, 300);

    return () => {
      clearTimeout(timeoutId);
      ignore = true;
    };
  }, [archiveFolderId, driveFolderId, isAddingSong, searchValue, t]);

  let status: React.ReactNode =
    searchResults.length === 1
      ? t("resultFound", { count: searchResults.length })
      : t("resultsFound", { count: searchResults.length });
  if (isAddingSong) {
    status = t("addingSong");
  } else if (isLoading) {
    status = (
      // todo: localize.
      <React.Fragment>
        <div
          className="size-4 animate-spin rounded-full border-2 border-gray-200 border-t-gray-600"
          aria-hidden
        />
        {t("searching")}
      </React.Fragment>
    );
  } else if (error) {
    status = error;
  } else if (searchResults.length === 0 && searchValue) {
    // todo: localize.
    status = t("nothingFound", { query: searchValue });
  }

  const shouldRenderPopup = searchValue !== "" || isAddingSong;

  return (
    <Autocomplete.Root
      open={shouldRenderPopup}
      items={searchResults}
      value={searchValue}
      onValueChange={setSearchValue}
      itemToStringValue={() => ""}
      filter={null}
      modal={true}
    >
      <div
        ref={wrapperRef}
        className="box-border flex min-h-[48px] items-center gap-[12px] rounded-[12px] bg-[var(--tgui--bg_color)] px-[16px] py-[12px]"
      >
        <Autocomplete.Input
          placeholder={t("nameOrLyricsPlaceholder")}
          className={
            "placeholder:text-darkgrey m-0 box-border block w-full resize-none border-0 bg-transparent p-0 font-[family-name:var(--tgui--font-family)] text-[length:var(--tgui--text--font_size)] leading-[var(--tgui--text--line_height)] font-[var(--tgui--font_weight--accent3)] overflow-ellipsis text-[var(--tgui--text_color)] outline-0"
          }
        />

        <IconButton mode="plain" size="s">
          <MagnifyingGlassIcon className="h-5 w-5 text-[var(--tg-theme-accent-text-color)]" />
        </IconButton>
      </div>

      {shouldRenderPopup && (
        <Autocomplete.Portal>
          <Autocomplete.Positioner
            anchor={wrapperRef}
            className="outline-none"
            sideOffset={8}
            align="start"
          >
            <Autocomplete.Popup
              className="box-border flex max-h-[min(var(--available-height),23rem)] w-[var(--anchor-width)] max-w-[var(--available-width)] scroll-pt-2 scroll-pb-2 flex-col overflow-y-auto overscroll-contain rounded-[12px] bg-[var(--tg-theme-bg-color)] text-[var(--tgui--text_color)] shadow-[0_32px_64px_0_rgba(0,0,0,0.04),_0_0_2px_1px_rgba(0,0,0,0.02)]"
              aria-busy={isLoading || undefined}
            >
              <Autocomplete.Status
                className={`flex items-center gap-2 pt-5 pr-8 pl-4 pb-${searchResults.length === 0 ? "5" : "2"} text-sm text-[var(--tg-theme-hint-color)]`}
              >
                {status}
              </Autocomplete.Status>
              <Autocomplete.List>
                {(driveFile: DriveFile) => (
                  <div>
                    <Autocomplete.Item
                      key={driveFile.id}
                      className="flex min-h-12 cursor-pointer items-center gap-4 px-4 py-4 hover:bg-[color-mix(in_srgb,var(--tg-theme-secondary-bg-color)_50%,transparent)] active:bg-[color-mix(in_srgb,var(--tg-theme-secondary-bg-color)_50%,transparent)]"
                      // className="flex h-12 cursor-pointer items-center gap-4 px-4 hover:bg-[var(--tgui--tertiary_bg_color)] hover:[opacity:0.85]"
                      // className="flex cursor-default py-2 pr-8 pl-4 text-base leading-4 outline-none select-none data-[highlighted]:relative data-[highlighted]:z-0 data-[highlighted]:text-gray-50 data-[highlighted]:before:absolute data-[highlighted]:before:inset-x-2 data-[highlighted]:before:inset-y-0 data-[highlighted]:before:z-[-1] data-[highlighted]:before:rounded data-[highlighted]:before:bg-gray-900"
                      value={driveFile}
                      onClick={() => {
                        hapticFeedback.impactOccurred("light");

                        setIsAddingSong(true);

                        getSongByDriveFileId(driveFile.id)
                          .then((data) => {
                            if (!data) {
                              throw new Error(
                                "Song not found from drive file id",
                              );
                            }

                            const song: Song = {
                              id: data.song.id,
                              name: data.song.pdf.name,
                              key: data.song.pdf.key,
                              bpm: data.song.pdf.bpm,
                              time: data.song.pdf.time,
                            };

                            onSelectSong?.(song);
                          })
                          .catch((err) => {
                            console.error(err);
                            Notify.failure(t("errorAddingSong"));
                            hapticFeedback.notificationOccurred("error");
                          })
                          .finally(() => {
                            setIsAddingSong(false);
                          });
                      }}
                    >
                      <div className="flex w-full flex-col gap-1">
                        <div className="leading-5 font-medium text-[var(--tg-theme-text-color)]">
                          {driveFile.name}
                        </div>
                        {/*<div className="text-sm leading-4 opacity-80">*/}
                        {/*  {driveFile.id}*/}
                        {/*</div>*/}
                      </div>
                    </Autocomplete.Item>
                  </div>
                )}
              </Autocomplete.List>
            </Autocomplete.Popup>
          </Autocomplete.Positioner>
        </Autocomplete.Portal>
      )}
    </Autocomplete.Root>
  );
}
