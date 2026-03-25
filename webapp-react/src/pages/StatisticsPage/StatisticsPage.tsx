import { getStatistics } from "@/api/webapp/statistics.ts";
import { StatisticsRole } from "@/api/webapp/typesResp.ts";
import { Page } from "@/components/Page.tsx";
import {
  buildStatisticsModel,
  StatisticsFilters,
} from "@/pages/StatisticsPage/statistics.ts";
import { useSuspenseQuery } from "@tanstack/react-query";
import {
  Accordion,
  Button,
  Chip,
  Input,
  List,
  Placeholder,
  Section,
  Spinner,
} from "@telegram-apps/telegram-ui";
import { hapticFeedback, postEvent } from "@tma.js/sdk-react";
import { FC, ReactNode, startTransition, useEffect, useMemo, useState } from "react";
import { useTranslation } from "react-i18next";
import { useSearchParams } from "react-router";

interface FilterChip<T extends string | number> {
  id: T;
  label: string;
}

export function StatisticsState({
  header,
  description,
  loading = false,
}: {
  header: string;
  description: string;
  loading?: boolean;
}) {
  return (
    <div className="px-3 py-10">
      <Placeholder
        header={header}
        description={description}
        className="rounded-[28px] bg-transparent"
      >
        {loading ? (
          <Spinner size="l" />
        ) : (
          <div
            aria-hidden
            className="h-4 w-4 rounded-full bg-[var(--tg-theme-button-color)] shadow-[0_0_0_8px_rgb(37_99_235_/_0.12)]"
          />
        )}
      </Placeholder>
    </div>
  );
}

function StatisticsChipGroup<T extends string | number>({
  items,
  selected,
  onToggle,
}: {
  items: FilterChip<T>[];
  selected: T[];
  onToggle: (value: T) => void;
}) {
  return (
    <div className="flex gap-2 overflow-x-auto px-1 py-1 [scrollbar-width:none] md:flex-wrap md:overflow-visible [&::-webkit-scrollbar]:hidden">
      {items.map((item) => {
        const isSelected = selected.includes(item.id);
        return (
          <Chip
            key={String(item.id)}
            Component="button"
            mode={isSelected ? "mono" : "outline"}
            aria-pressed={isSelected}
            className={`shrink-0 transition-transform hover:-translate-y-px ${
              isSelected
                ? "!border-transparent !bg-[var(--tg-theme-button-color)] !text-[var(--tg-theme-button-text-color)] shadow-none"
                : "!bg-[var(--tg-theme-section-bg-color)] !text-[var(--tg-theme-text-color)] shadow-[0_0_0_1px_var(--tg-theme-section-separator-color)]"
            }`}
            onClick={() => {
              hapticFeedback.selectionChanged();
              startTransition(() => {
                onToggle(item.id);
              });
            }}
          >
            {item.label}
          </Chip>
        );
      })}
    </div>
  );
}

function toggleValue<T extends string | number>(values: T[], value: T): T[] {
  return values.includes(value)
    ? values.filter((currentValue) => currentValue !== value)
    : [...values, value];
}

function formatDisplayDate(value: string, language: string): string {
  return new Intl.DateTimeFormat(language, {
    day: "numeric",
    month: "short",
    year: "numeric",
  }).format(new Date(`${value}T00:00:00`));
}

function parseISODate(value: string): Date {
  const [year, month, day] = value.split("-").map(Number);
  return new Date(year, month - 1, day);
}

function formatISODate(value: Date): string {
  const year = value.getFullYear();
  const month = String(value.getMonth() + 1).padStart(2, "0");
  const day = String(value.getDate()).padStart(2, "0");
  return `${year}-${month}-${day}`;
}

function addMonthsToISODate(value: string, deltaMonths: number): string {
  const source = parseISODate(value);
  const target = new Date(source.getFullYear(), source.getMonth() + deltaMonths, 1);
  const lastDayOfTargetMonth = new Date(
    target.getFullYear(),
    target.getMonth() + 1,
    0,
  ).getDate();

  target.setDate(Math.min(source.getDate(), lastDayOfTargetMonth));
  return formatISODate(target);
}

function startOfISODateYear(value: string): string {
  const source = parseISODate(value);
  return `${source.getFullYear()}-01-01`;
}

function renderRoleBreakdown(
  items: Array<{ roleId: string; roleName: string; count: number }>,
): ReactNode {
  return items.map((role) => (
    <Chip
      key={role.roleId}
      mode="outline"
      className="!bg-transparent text-[var(--tg-theme-text-color)]"
    >
      <span className="mr-2 font-roboto-mono text-sm font-semibold">
        {role.count}
      </span>
      <span>{role.roleName}</span>
    </Chip>
  ));
}

function AdvancedFiltersContent({
  roleOptions,
  weekdayOptions,
  selectedRoleIds,
  selectedWeekdays,
  onRoleToggle,
  onWeekdayToggle,
  t,
}: {
  roleOptions: FilterChip<string>[];
  weekdayOptions: FilterChip<number>[];
  selectedRoleIds: string[];
  selectedWeekdays: number[];
  onRoleToggle: (value: string) => void;
  onWeekdayToggle: (value: number) => void;
  t: (key: string, options?: Record<string, unknown>) => string;
}) {
  return (
    <div className="grid gap-4">
      {roleOptions.length > 0 ? (
        <div className="grid gap-2">
          <p className="text-xs font-semibold tracking-[0.08em] text-[var(--tg-theme-hint-color)] uppercase">
            {t("statisticsRoles")}
          </p>
          <StatisticsChipGroup
            items={roleOptions}
            selected={selectedRoleIds}
            onToggle={onRoleToggle}
          />
        </div>
      ) : null}

      <div className="grid gap-2">
        <p className="text-xs font-semibold tracking-[0.08em] text-[var(--tg-theme-hint-color)] uppercase">
          {t("statisticsWeekdays")}
        </p>
        <StatisticsChipGroup
          items={weekdayOptions}
          selected={selectedWeekdays}
          onToggle={onWeekdayToggle}
        />
      </div>
    </div>
  );
}

const StatisticsPage: FC = () => {
  const { i18n, t } = useTranslation();
  const [searchParams] = useSearchParams();
  const bandId = searchParams.get("bandId");

  if (!bandId) {
    throw new Error("Failed to get statistics page: invalid request params.");
  }

  const [selectedFromDate, setSelectedFromDate] = useState("");
  const [selectedRoleIds, setSelectedRoleIds] = useState<string[]>([]);
  const [selectedWeekdays, setSelectedWeekdays] = useState<number[]>([]);
  const [isAdvancedFiltersOpen, setIsAdvancedFiltersOpen] = useState(false);

  useEffect(() => {
    postEvent("web_app_expand");
    postEvent("web_app_setup_swipe_behavior", { allow_vertical_swipe: true });
  }, []);

  const query = useSuspenseQuery({
    queryKey: ["statistics", bandId, selectedFromDate || "default"],
    queryFn: async () => {
      const data = await getStatistics(bandId, selectedFromDate || undefined);
      if (!data) {
        throw new Error("Failed to get statistics data.");
      }
      return data;
    },
  });

  if (query.isError || !query.data) {
    return (
      <Page back={false}>
        <StatisticsState
          header={t("statisticsErrorTitle")}
          description={t("statisticsErrorDescription")}
        />
      </Page>
    );
  }

  const effectiveFromDate = selectedFromDate || query.data.defaultFromDate;
  const filters: StatisticsFilters = useMemo(() => ({
    fromDate: effectiveFromDate,
    selectedRoleIds,
    selectedWeekdays,
  }), [effectiveFromDate, selectedRoleIds, selectedWeekdays]);

  const statistics = useMemo(() => buildStatisticsModel(
    query.data.users,
    query.data.roles,
    filters,
  ), [query.data.users, query.data.roles, filters]);

  const weekdayOptions: FilterChip<number>[] = useMemo(() => [
    { id: 1, label: t("weekdayMonShort") },
    { id: 2, label: t("weekdayTueShort") },
    { id: 3, label: t("weekdayWedShort") },
    { id: 4, label: t("weekdayThuShort") },
    { id: 5, label: t("weekdayFriShort") },
    { id: 6, label: t("weekdaySatShort") },
    { id: 0, label: t("weekdaySunShort") },
  ], [t]);

  const roleOptions: FilterChip<string>[] = useMemo(() => (query.data.roles ?? []).map(
    (role: StatisticsRole) => ({
      id: role.id,
      label: role.name,
    }),
  ), [query.data.roles]);
  const datePresetOptions: FilterChip<string>[] = useMemo(() => [
    {
      id: addMonthsToISODate(query.data.currentDate, -1),
      label: t("statisticsPresetLastMonth"),
    },
    {
      id: addMonthsToISODate(query.data.currentDate, -6),
      label: t("statisticsPresetLastSixMonths"),
    },
    {
      id: addMonthsToISODate(query.data.currentDate, -12),
      label: t("statisticsPresetLastYear"),
    },
    {
      id: startOfISODateYear(query.data.currentDate),
      label: t("statisticsPresetYearStart"),
    },
  ], [query.data.currentDate, t]);
  const activeDatePresetIds = datePresetOptions.some(
    ({ id }) => id === effectiveFromDate,
  )
    ? [effectiveFromDate]
    : [];

  const applyFromDate = (nextValue: string) => {
    const normalizedValue =
      nextValue === query.data.defaultFromDate ? "" : nextValue;

    startTransition(() => {
      setSelectedFromDate(normalizedValue);
    });
  };

  const hasActiveFilters =
    selectedFromDate.length > 0 ||
    selectedRoleIds.length > 0 ||
    selectedWeekdays.length > 0;
  return (
    <Page back={false}>
      <div className="min-h-screen text-[var(--tg-theme-text-color)]">
        <div className="mx-auto flex max-w-5xl flex-col gap-4 px-0 pt-6 pb-[calc(2.5rem+var(--tg-safe-area-inset-bottom,0px))]">
          <List className="bg-transparent">
            <Section
              header={
                <Section.Header large>{query.data.bandName}</Section.Header>
              }
              className="overflow-hidden"
            >
              <div className="grid gap-4 px-4 py-4 md:px-5">
                <div className="flex flex-col gap-2 md:flex-row md:items-start md:justify-between">
                  <div className="space-y-2">
                    <h1 className="text-[clamp(1.85rem,5vw,3.1rem)] leading-none font-semibold tracking-[-0.04em]">
                      {t("statisticsTitle")}
                    </h1>
                    <p className="max-w-2xl text-sm leading-5 text-[var(--tg-theme-hint-color)] md:text-[0.95rem]">
                      {t("statisticsSubtitle", {
                        date: formatDisplayDate(
                          effectiveFromDate,
                          i18n.language,
                        ),
                      })}
                    </p>
                  </div>
                </div>

                <div className="grid grid-cols-2 gap-3 border-t border-[color:var(--tg-theme-section-separator-color)] pt-3">
                  <div className="flex h-full flex-col justify-between gap-2">
                    <p className="text-xs font-semibold tracking-[0.08em] text-[var(--tg-theme-hint-color)] uppercase">
                      {t("statisticsTotalParticipations")}
                    </p>
                    <p className="font-roboto-mono text-[clamp(1.95rem,8vw,3.1rem)] leading-none tracking-[-0.06em]">
                      {statistics.totalParticipations}
                    </p>
                  </div>

                  <div className="flex h-full flex-col justify-between gap-2 border-l border-[color:var(--tg-theme-section-separator-color)] pl-3">
                    <p className="text-xs font-semibold tracking-[0.08em] text-[var(--tg-theme-hint-color)] uppercase">
                      {t("statisticsActiveMembers")}
                    </p>
                    <p className="font-roboto-mono text-[clamp(1.95rem,8vw,3.1rem)] leading-none tracking-[-0.06em]">
                      {statistics.activeMembers}
                    </p>
                  </div>
                </div>
              </div>
            </Section>

            <Section className="overflow-hidden">
              <div className="grid gap-4 px-4 py-4 md:px-5">
                <div className="flex flex-col gap-3">
                  <div className="grid w-full gap-2 md:max-w-[18rem]">
                    <p className="text-xs font-semibold tracking-[0.08em] text-[var(--tg-theme-hint-color)] uppercase">
                      {t("statisticsFromDate")}
                    </p>
                    <Input
                      type="date"
                      className="w-full rounded-[20px] !bg-inherit px-3 py-2 text-base font-medium shadow-[0_0_0_1px_var(--tgui--outline)]"
                      value={effectiveFromDate}
                      max={query.data.currentDate}
                      onChange={(event) => {
                        const nextValue =
                          event.target.value > query.data.currentDate
                            ? query.data.currentDate
                            : event.target.value;
                        hapticFeedback.selectionChanged();
                        applyFromDate(nextValue);
                      }}
                    />
                  </div>

                  <StatisticsChipGroup
                    items={datePresetOptions}
                    selected={activeDatePresetIds}
                    onToggle={(value) => {
                      applyFromDate(value);
                    }}
                  />
                </div>

                <div className="hidden md:block">
                  <AdvancedFiltersContent
                    roleOptions={roleOptions}
                    weekdayOptions={weekdayOptions}
                    selectedRoleIds={selectedRoleIds}
                    selectedWeekdays={selectedWeekdays}
                    onRoleToggle={(value) => {
                      setSelectedRoleIds((currentValues) =>
                        toggleValue(currentValues, value),
                      );
                    }}
                    onWeekdayToggle={(value) => {
                      setSelectedWeekdays((currentValues) =>
                        toggleValue(currentValues, value),
                      );
                    }}
                    t={t}
                  />
                </div>

                <div className="md:hidden">
                  <Accordion
                    expanded={isAdvancedFiltersOpen}
                    onChange={(nextExpanded) => {
                      hapticFeedback.impactOccurred("light");
                      setIsAdvancedFiltersOpen(nextExpanded);
                    }}
                  >
                    <Accordion.Summary
                      className="hover:!bg-transparent active:!bg-transparent"
                      multiline
                    >
                      {t("statisticsFilters")}
                    </Accordion.Summary>

                    <Accordion.Content
                      className="overflow-hidden rounded-b-[18px]"
                      style={{
                        backgroundColor: "var(--tg-theme-section-bg-color)",
                      }}
                    >
                      <div
                        className="pt-2 pb-4"
                        style={{
                          backgroundColor: "var(--tg-theme-section-bg-color)",
                        }}
                      >
                        <div
                          className="rounded-2xl px-3 py-3"
                          style={{
                            backgroundColor:
                              "var(--tg-theme-secondary-bg-color)",
                          }}
                        >
                          <AdvancedFiltersContent
                            roleOptions={roleOptions}
                            weekdayOptions={weekdayOptions}
                            selectedRoleIds={selectedRoleIds}
                            selectedWeekdays={selectedWeekdays}
                            onRoleToggle={(value) => {
                              setSelectedRoleIds((currentValues) =>
                                toggleValue(currentValues, value),
                              );
                            }}
                            onWeekdayToggle={(value) => {
                              setSelectedWeekdays((currentValues) =>
                                toggleValue(currentValues, value),
                              );
                            }}
                            t={t}
                          />
                        </div>
                      </div>
                    </Accordion.Content>
                  </Accordion>
                </div>

                {hasActiveFilters ? (
                  <div className="flex justify-start md:justify-end">
                    <Button
                      mode="outline"
                      size="s"
                      onClick={() => {
                        hapticFeedback.impactOccurred("light");
                        startTransition(() => {
                          setSelectedFromDate("");
                          setSelectedRoleIds([]);
                          setSelectedWeekdays([]);
                          setIsAdvancedFiltersOpen(false);
                        });
                      }}
                    >
                      {t("statisticsClearFilters")}
                    </Button>
                  </div>
                ) : null}
              </div>
            </Section>

            <Section className="overflow-hidden">
              {statistics.rows.length === 0 ? (
                <StatisticsState
                  header={t("statisticsEmptyTitle")}
                  description={t("statisticsEmptyDescription")}
                />
              ) : (
                <div className="px-4 py-3 md:px-5">
                  <div className="hidden grid-cols-[minmax(0,1.6fr)_minmax(0,2fr)_auto] gap-4 pb-3 text-xs font-semibold tracking-[0.08em] text-[var(--tg-theme-hint-color)] uppercase md:grid">
                    <span>{t("statisticsMemberColumn")}</span>
                    <span>{t("statisticsRoleBreakdownColumn")}</span>
                    <span className="text-right">
                      {t("statisticsCountColumn")}
                    </span>
                  </div>

                  <div className="divide-y divide-[color:var(--tg-theme-section-separator-color)]">
                    {statistics.rows.map((row, index) => (
                      <article
                        key={row.userId}
                        className="grid gap-4 py-4 md:grid-cols-[minmax(0,1.6fr)_minmax(0,2fr)_auto] md:items-start"
                      >
                        <div className="grid grid-cols-[auto_1fr] gap-3">
                          <span className="pt-0.5 font-roboto-mono text-sm text-[var(--tg-theme-hint-color)]">
                            {String(index + 1).padStart(2, "0")}
                          </span>
                          <div>
                            <h2 className="text-base leading-5 font-medium">
                              {row.name}
                            </h2>
                            <p className="mt-1 text-sm leading-5 text-[var(--tg-theme-hint-color)]">
                              {row.lastEventDate
                                ? t("statisticsLastEvent", {
                                    date: formatDisplayDate(
                                      row.lastEventDate,
                                      i18n.language,
                                    ),
                                  })
                                : t("statisticsNoRecentEvents")}
                            </p>
                          </div>
                        </div>

                        <div className="flex flex-wrap gap-2 pl-8 md:pl-0">
                          {row.roleBreakdown.length > 0 ? (
                            renderRoleBreakdown(row.roleBreakdown)
                          ) : (
                            <span className="text-sm text-[var(--tg-theme-hint-color)]">
                              {t("statisticsNoRoles")}
                            </span>
                          )}
                        </div>

                        <div className="flex items-end justify-between pl-8 md:flex-col md:items-end md:pl-0">
                          <span className="font-roboto-mono text-[2rem] leading-none">
                            {row.participationCount}
                          </span>
                          <span className="text-[0.72rem] tracking-[0.08em] text-[var(--tg-theme-hint-color)] uppercase">
                            {t("statisticsCountLabel")}
                          </span>
                        </div>
                      </article>
                    ))}
                  </div>
                </div>
              )}
            </Section>
          </List>
        </div>
      </div>
    </Page>
  );
};

export default StatisticsPage;
