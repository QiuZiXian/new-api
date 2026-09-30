/*
Copyright (C) 2023-2026 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.

For commercial licensing, please contact support@quantumnous.com
*/
import { Code2, PenTool, Building2 } from "lucide-react";
import { useTranslation } from "react-i18next";

import { AnimateInView } from "@/components/animate-in-view";

export function HowItWorks() {
  const { t } = useTranslation();

  const teams = [
    {
      icon: <Code2 className="size-5" strokeWidth={1.6} />,
      title: t("Developer teams"),
      desc: t(
        "Integrate multi-model capabilities quickly and cut vendor adaptation and API maintenance costs.",
      ),
      tags: [t("API access"), t("Model switching"), t("Call logs")],
    },
    {
      icon: <PenTool className="size-5" strokeWidth={1.6} />,
      title: t("Content teams"),
      desc: t(
        "Call image, video and speech models in one place to speed up marketing assets and short-video production.",
      ),
      tags: [
        t("Image generation"),
        t("Video generation"),
        t("Speech generation"),
      ],
    },
    {
      icon: <Building2 className="size-5" strokeWidth={1.6} />,
      title: t("Enterprise admins"),
      desc: t(
        "Review usage, cost, key permissions and call records centrally across teams with less overhead.",
      ),
      tags: [t("Cost management"), t("Permission control"), t("Usage stats")],
    },
  ];

  return (
    <section
      aria-labelledby="how-it-works-heading"
      className="relative z-10 px-4 py-8 md:px-6 md:py-10"
    >
      {/* Background using the left half of the design asset */}
      <div className="relative isolate mx-auto max-w-[1400px] overflow-hidden rounded-[24px] border border-indigo-100/80 px-6 py-12 shadow-[0_30px_70px_-52px_rgba(79,70,229,0.5)] md:rounded-[32px] md:px-12 md:py-18 dark:border-white/10">
        <img
          aria-hidden
          src="/home/hero-slide-1.jpg"
          alt=""
          className="absolute inset-y-0 left-0 -z-30 w-1/2 object-cover object-left"
        />
        <div
          aria-hidden
          className="absolute inset-0 -z-20 bg-white/30 dark:bg-slate-900/60"
        />

        <div className="mx-auto max-w-6xl">
          <AnimateInView className="mb-12 text-center md:mb-14">
            <span className="inline-flex items-center rounded-full border border-indigo-200/80 bg-white/70 px-3 py-1 text-[11px] font-medium tracking-[0.18em] text-indigo-600 uppercase backdrop-blur-sm dark:border-white/15 dark:bg-white/10 dark:text-indigo-300">
              {t("Teams")}
            </span>
            <h2
              id="how-it-works-heading"
              className="mt-5 text-[clamp(1.75rem,3vw,2.5rem)] leading-[1.18] font-bold tracking-tight text-slate-900 dark:text-slate-50"
            >
              {t("Built for every AI-powered team")}
            </h2>
            <p className="mx-auto mt-4 max-w-2xl text-[15px] leading-relaxed text-slate-500 dark:text-slate-400">
              {t(
                "Whether it is API integration, content production or enterprise governance, one platform covers the workflow.",
              )}
            </p>
          </AnimateInView>

          <div className="grid gap-5 md:grid-cols-3">
            {teams.map((team, i) => (
              <AnimateInView
                key={team.title}
                delay={i * 130}
                animation="fade-up"
                className="group relative flex flex-col overflow-hidden rounded-[20px] border border-slate-200/70 bg-white/85 p-6 shadow-[0_18px_40px_-30px_rgba(15,23,42,0.35)] backdrop-blur-sm transition-all duration-300 motion-safe:hover:-translate-y-1 hover:border-indigo-300 hover:shadow-[0_28px_60px_-32px_rgba(79,70,229,0.45)] md:p-7 dark:border-white/10 dark:bg-white/5"
              >
                {/* Hover accent bar */}
                <span
                  aria-hidden
                  className="absolute inset-x-0 top-0 h-[3px] bg-gradient-to-r from-blue-500 via-indigo-500 to-violet-500 opacity-0 transition-opacity duration-300 group-hover:opacity-100"
                />
                <span className="mb-5 inline-flex size-12 items-center justify-center rounded-[14px] bg-gradient-to-br from-blue-600 to-indigo-600 text-white shadow-[0_10px_22px_-12px_rgba(37,99,235,0.7)]">
                  {team.icon}
                </span>
                <h3 className="mb-2 text-[17px] font-semibold text-slate-900 dark:text-slate-50">
                  {team.title}
                </h3>
                <p className="mb-6 text-sm leading-relaxed text-slate-500 dark:text-slate-400">
                  {team.desc}
                </p>
                <div className="mt-auto flex flex-wrap gap-1.5">
                  {team.tags.map((tag) => (
                    <span
                      key={tag}
                      className="rounded-full border border-blue-100 bg-blue-50 px-2.5 py-1 text-[11px] font-medium text-blue-700 transition-colors group-hover:border-blue-200 group-hover:bg-blue-100/70 dark:border-blue-400/25 dark:bg-blue-500/10 dark:text-blue-300"
                    >
                      {tag}
                    </span>
                  ))}
                </div>
              </AnimateInView>
            ))}
          </div>
        </div>
      </div>
    </section>
  );
}
