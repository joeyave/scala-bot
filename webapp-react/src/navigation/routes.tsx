import type { ComponentType, JSX } from "react";

import { IndexPage } from "@/pages/IndexPage/IndexPage";
import { InitDataPage } from "@/pages/InitDataPage.tsx";
import { LaunchParamsPage } from "@/pages/LaunchParamsPage.tsx";
import { SongConfirmationPage } from "@/pages/SongPage/SongConfirmationPage.tsx";
import { SongPage } from "@/pages/SongPage/SongPage.tsx";
import { ThemeParamsPage } from "@/pages/ThemeParamsPage.tsx";

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

  { path: "/songs/:id/edit", Component: SongPage },
  {
    path: "/songs/:id/edit/confirm",
    Component: SongConfirmationPage,
  },
];
