import { ComponentType, JSX, lazy } from "react";

import { IndexPage } from "@/pages/IndexPage/IndexPage";
import { InitDataPage } from "@/pages/InitDataPage.tsx";
import { LaunchParamsPage } from "@/pages/LaunchParamsPage.tsx";
import { ThemeParamsPage } from "@/pages/ThemeParamsPage.tsx";

const SongPage = lazy(() => import("@/pages/SongPage/SongPage"));
const CreateSongPage = lazy(() => import("@/pages/SongPage/CreateSongPage"));
const SongConfirmationPage = lazy(
  () => import("@/pages/SongPage/SongConfirmationPage"),
);

interface Route {
  path: string;
  Component: ComponentType;
  title?: string;
  icon?: JSX.Element;
}

export const routes: Route[] = [
  { path: "/", Component: IndexPage },
  { path: "/init-data", Component: InitDataPage, title: "Init Data" },
  { path: "/theme-params", Component: ThemeParamsPage, title: "Theme Params" },
  {
    path: "/launch-params",
    Component: LaunchParamsPage,
    title: "Launch Params",
  },

  {
    path: "/songs/create",
    Component: CreateSongPage,
  },
  { path: "/songs/:songId/edit", Component: SongPage },
  {
    path: "/songs/:songId/edit/confirm",
    Component: SongConfirmationPage,
  },
];
