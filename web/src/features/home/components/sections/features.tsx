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
import { KeyRound, Route, Wallet, ShieldCheck } from "lucide-react";
import { useTranslation } from "react-i18next";

import { AnimateInView } from "@/components/animate-in-view";

interface FeaturesProps {
  className?: string;
}

export function Features(_props: FeaturesProps) {
  const { t } = useTranslation();

  const capabilities = [
    {
      icon: <KeyRound className="size-6" strokeWidth={1.8} />,
      title: t("Unified access"),
      desc: t(
        "One API key calls text, image, video and multimodal models, cutting repeated vendor integration costs.",
      ),
      iconClass:
        "bg-blue-500/15 text-blue-600 dark:bg-blue-400/15 dark:text-blue-400",
    },
    {
      icon: <Wallet className="size-6" strokeWidth={1.8} />,
      title: t("Predictable cost"),
      desc: t(
        "Pay-as-you-go and per-request billing with quota limits and usage stats keep every call accountable.",
      ),
      iconClass:
        "bg-amber-500/15 text-amber-600 dark:bg-amber-400/15 dark:text-amber-400",
    },
    {
      icon: <ShieldCheck className="size-6" strokeWidth={1.8} />,
      title: t("Permission control"),
      desc: t(
        "Key-level quotas, expiry, model restrictions and IP allowlists fit per-project and per-team governance.",
      ),
      iconClass:
        "bg-violet-500/15 text-violet-600 dark:bg-violet-400/15 dark:text-violet-400",
    },
    {
      icon: <Route className="size-6" strokeWidth={1.8} />,
      title: t("Stable routing"),
      desc: t(
        "Unified scheduling across multiple channels reduces the impact of a single model or account outage.",
      ),
      iconClass:
        "bg-emerald-500/15 text-emerald-600 dark:bg-emerald-400/15 dark:text-emerald-400",
    },
  ];

  return (
    <section
      aria-labelledby="features-heading"
      className="relative z-10 px-4 py-8 md:px-6 md:py-10"
    >
      {/* Background fills the panel with the left half of the design asset */}
      <div className="relative isolate mx-auto max-w-[1400px] overflow-hidden rounded-[24px] border border-indigo-100/80 px-6 py-12 shadow-[0_30px_70px_-52px_rgba(79,70,229,0.5)] md:rounded-[32px] md:px-12 md:py-18 dark:border-white/10">
        <img
          aria-hidden
          src="/home/hero-slide-2.jpg"
          alt=""
          className="absolute inset-0 -z-10 h-full w-full object-cover object-left"
        />

        <div className="mx-auto max-w-6xl">
          <AnimateInView className="mb-12 text-center md:mb-14">
            <span className="inline-flex items-center rounded-full border border-indigo-200/80 bg-white/70 px-3 py-1 text-[11px] font-medium tracking-[0.18em] text-indigo-600 uppercase backdrop-blur-sm dark:border-white/15 dark:bg-white/10 dark:text-indigo-300">
              {t("Enterprise Capabilities")}
            </span>
            <h2
              id="features-heading"
              className="mt-5 text-[clamp(1.75rem,3vw,2.5rem)] leading-[1.18] font-bold tracking-tight"
            >
              <span className="bg-gradient-to-r from-blue-600 via-indigo-500 to-violet-500 bg-clip-text text-transparent dark:from-blue-400 dark:via-indigo-400 dark:to-violet-400">
                {t("Designed for enterprise-grade AI integration")}
              </span>
            </h2>
            <p className="mx-auto mt-4 max-w-2xl text-[15px] leading-relaxed text-slate-500 dark:text-slate-400">
              {t(
                "From model calls to usage management, teams get unified access, unified billing and unified governance.",
              )}
            </p>
          </AnimateInView>

          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
            {capabilities.map((capability, i) => (
              <AnimateInView
                key={capability.title}
                delay={i * 110}
                animation="fade-up"
                className="group relative flex flex-col overflow-hidden rounded-[20px] border border-slate-200/70 bg-white/85 p-6 shadow-[0_18px_40px_-30px_rgba(15,23,42,0.35)] backdrop-blur-sm transition-all duration-300 motion-safe:hover:-translate-y-1 hover:border-indigo-300 hover:shadow-[0_28px_60px_-32px_rgba(79,70,229,0.45)] dark:border-white/10 dark:bg-white/5"
              >
                {/* Hover accent bar */}
                <span
                  aria-hidden
                  className="absolute inset-x-0 top-0 h-[3px] bg-gradient-to-r from-blue-500 via-indigo-500 to-violet-500 opacity-0 transition-opacity duration-300 group-hover:opacity-100"
                />
                {/* Index */}
                <span className="absolute top-5 right-5 text-[13px] font-semibold tabular-nums text-slate-300 transition-colors group-hover:text-slate-400 dark:text-white/25 dark:group-hover:text-white/50">
                  {String(i + 1).padStart(2, "0")}
                </span>

                <span
                  className={`mb-5 inline-flex size-12 items-center justify-center rounded-[14px] transition-colors duration-300 ${capability.iconClass}`}
                >
                  {capability.icon}
                </span>
                <h3 className="mb-2 text-[17px] font-semibold text-slate-900 dark:text-slate-50">
                  {capability.title}
                </h3>
                <p className="text-sm leading-relaxed text-slate-500 dark:text-slate-400">
                  {capability.desc}
                </p>
              </AnimateInView>
            ))}
          </div>
        </div>
      </div>
    </section>
  );
}
