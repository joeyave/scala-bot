import { updateSong } from "@/api/webapp/songs.ts";
import {
  ReqBodyUpdateSong,
  ReqQueryParamsUpdateSong,
} from "@/api/webapp/typesReq.ts";
import { Page } from "@/components/Page.tsx";
import { logger } from "@/helpers/logger";
import { PageError } from "@/pages/UtilPages/PageError.tsx";
import { backButton, mainButton, miniApp } from "@telegram-apps/sdk-react";
import {
  List,
  Section,
  Select,
  Subheadline,
  Text,
} from "@telegram-apps/telegram-ui";
import { FC, useEffect, useState } from "react";
import { useLocation, useNavigate } from "react-router-dom";

import { SongFormValues } from "@/pages/SongPage/util/types.ts";
import { Location } from "react-router";

export interface LocationWithState extends Location {
  state: {
    songId: string;
    userId: string;
    chatId: string;
    messageId: string;
    formData: SongFormValues;
    initialFormData: SongFormValues;
    sectionsNumber: number;
  };
}

export const useAppLocation = (): LocationWithState =>
  useLocation() as LocationWithState;

export const SongConfirmationPage: FC = () => {
  const navigate = useNavigate();
  const [error, setError] = useState<Error | null>(null);
  const [sectionNumber, setSectionNumber] = useState<string>("-1");

  const { state } = useAppLocation();

  const {
    songId,
    userId,
    chatId,
    messageId,
    formData,
    initialFormData,
    sectionsNumber,
  } = state || {};

  useEffect(() => {
    // Show main button for final save
    mainButton.setParams({
      isVisible: true,
      isEnabled: true,
      text: "Save Changes",
    });

    // Handle main button click

    const handleMainButtonClick = async () => {
      backButton.hide();
      mainButton.setParams({ isLoaderVisible: true });

      const queryParams: ReqQueryParamsUpdateSong = {
        messageId: messageId,
        chatId: chatId,
        userId: userId,
      };
      // Prepare the form data to send
      const body: ReqBodyUpdateSong = {
        name: formData.name,
        key: formData.key || "?", // todo: refactor, remove question marks.
        bpm: formData.bpm || "?",
        time: formData.time || "?",
        tags: formData.tags.map((tag) => String(tag.value)),

        transposeSection: sectionNumber,
      };

      const err = await updateSong(songId, queryParams, body);
      if (err) {
        logger.error("Failed to save song", { error: err });
        setError(err);
        mainButton.setParams({ isVisible: false, isLoaderVisible: false });
        backButton.show();
      } else {
        miniApp.close();
      }
    };

    const handleMainButtonClickSync = () => {
      void handleMainButtonClick();
    };

    mainButton.onClick(handleMainButtonClickSync);

    const handleBackButtonClick = () => {
      navigate(`/songs/${songId}`);
    };
    backButton.onClick(handleBackButtonClick);

    return () => {
      mainButton.offClick(handleMainButtonClickSync);
      backButton.offClick(handleBackButtonClick);
    };
  }, [
    chatId,
    formData.bpm,
    formData.key,
    formData.name,
    formData.tags,
    formData.time,
    messageId,
    navigate,
    sectionNumber,
    songId,
    userId,
  ]);

  // Show error state
  if (error) {
    return <PageError error={error}></PageError>;
  }

  return (
    <Page back={true}>
      <List>
        <Subheadline>
          You are changing the key from {initialFormData.key} to {formData.key}.
          Where should page with new key be placed?
        </Subheadline>
        <Select
          onChange={(e) => {
            setSectionNumber(e.target.value);
          }}
        >
          <option key={"-1"} value={"-1"}>
            At the end of the doc
          </option>
          {Array(sectionsNumber)
            .fill(null)
            .map((_, i) => (
              <option key={String(i)} value={String(i)}>
                Instead of {i + 1} page
              </option>
            ))}
          <option key={""} value={""}>
            Only first header
          </option>
        </Select>

        <Section header="Song Information">
          <div style={{ padding: "16px" }}>
            {" "}
            {/*todo*/}
            <Text>Here goes PDF.</Text>
          </div>
        </Section>
      </List>
    </Page>
  );
};
