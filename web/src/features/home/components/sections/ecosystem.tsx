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
import { ArrowRight, Rocket, Sparkles, Tags, UserPlus } from "lucide-react";
import { useTranslation } from "react-i18next";

import { AnimateInView } from "@/components/animate-in-view";
import { Button } from "@/components/ui/button";
import { Link } from "@tanstack/react-router";

/**
 * Second screen — the "one gateway" value proposition with the three primary
 * entry actions. Deliberately light: it acts as a breathing space between the
 * hero and the deep-gradient capability banner that follows.
 */
export function Ecosystem() {
  const { t } = useTranslation();
  const isAuthenticated = false;

  return (
    <section
      aria-labelledby="ecosystem-heading"
      className="relative z-10 px-4 py-8 md:px-6 md:py-10"
    >
      <div className="relative isolate mx-auto max-w-[1400px] overflow-hidden rounded-[24px] border border-blue-100/80 px-6 py-12 shadow-[0_30px_70px_-52px_rgba(37,99,235,0.55)] md:rounded-[32px] md:px-12 md:py-16 dark:border-white/10">
        {/* Light blue panel base */}
        <div
          aria-hidden
          className="absolute inset-0 -z-20 bg-[linear-gradient(135deg,#f9fbff_0%,#eaf2ff_48%,#f6f9ff_100%)] dark:bg-[linear-gradient(135deg,#0a1830_0%,#122242_50%,#0a1830_100%)]"
        />
        {/* Fine grid texture, faded towards the edges */}
        <div
          aria-hidden
          className="absolute inset-0 -z-10 bg-[linear-gradient(to_right,rgba(37,99,235,0.07)_1px,transparent_1px),linear-gradient(to_bottom,rgba(37,99,235,0.07)_1px,transparent_1px)] bg-[size:44px_44px] [mask-image:radial-gradient(ellipse_75%_70%_at_50%_35%,black_25%,transparent_100%)]"
        />
        {/* Glows */}
        <div
          aria-hidden
          className="absolute -top-28 left-1/2 -z-10 size-[420px] -translate-x-1/2 rounded-full bg-blue-300/30 blur-2xl dark:bg-blue-500/15"
        />
        <div
          aria-hidden
          className="absolute -right-24 -bottom-24 -z-10 size-72 rounded-full bg-indigo-200/30 blur-2xl dark:bg-indigo-500/10"
        />

        <div className="mx-auto max-w-3xl text-center">
          <AnimateInView animation="fade-in">
            <span className="inline-flex items-center gap-1.5 rounded-full border border-blue-200/70 bg-white/70 px-3 py-1 text-[11px] font-medium tracking-[0.18em] text-blue-600 uppercase backdrop-blur-sm dark:border-white/15 dark:bg-white/10 dark:text-blue-300">
              <Sparkles className="size-3" />
              {t("Application Ecosystem")}
            </span>
          </AnimateInView>

          <AnimateInView delay={80}>
            <h2
              id="ecosystem-heading"
              className="mt-5 text-[clamp(1.75rem,3.2vw,2.5rem)] leading-[1.18] font-bold tracking-tight text-slate-900 dark:text-slate-50"
            >
              {t("One gateway, connected to")}
              <br />
              <span className="bg-gradient-to-r from-blue-600 via-indigo-500 to-violet-500 bg-clip-text text-transparent dark:from-blue-400 dark:via-indigo-400 dark:to-violet-400">
                {t("the tools you already use")}
              </span>
            </h2>
          </AnimateInView>

          <AnimateInView delay={140}>
            <p className="mx-auto mt-4 max-w-xl text-[15px] leading-relaxed text-slate-500 dark:text-slate-400">
              {t(
                "One gateway for every model — no changes to your existing apps.",
              )}
            </p>
          </AnimateInView>
        </div>

        <AnimateInView
          className="mt-10 flex flex-col items-center justify-center gap-3.5 sm:flex-row sm:gap-4"
          animation="fade-up"
          delay={200}
        >
          <Button
            className="h-[54px] w-full rounded-xl bg-gradient-to-r from-blue-600 to-blue-500 px-7 text-[15px] font-medium text-white shadow-[0_8px_24px_-6px_rgba(37,99,235,0.55)] transition-all duration-200 hover:from-blue-700 hover:to-blue-600 hover:shadow-[0_12px_28px_-8px_rgba(37,99,235,0.6)] focus-visible:outline-none focus-visible:ring-3 focus-visible:ring-blue-500/40 active:translate-y-px motion-safe:hover:-translate-y-0.5 sm:w-auto"
            render={
              isAuthenticated ? (
                <Link to="/dashboard" />
              ) : (
                <Link to="/sign-up" />
              )
            }
          >
            <UserPlus className="mr-2 size-[18px]" />
            {t("Register now")}
            <ArrowRight className="ml-2 size-4 transition-transform duration-300 group-hover/button:translate-x-0.5" />
          </Button>
          <Button
            className="h-[54px] w-full rounded-xl border-blue-200 bg-white px-7 text-[15px] font-medium text-blue-700 shadow-[0_8px_20px_-14px_rgba(37,99,235,0.45)] transition-all duration-200 hover:border-blue-300 hover:bg-blue-50/70 hover:shadow-[0_12px_24px_-12px_rgba(37,99,235,0.5)] focus-visible:outline-none focus-visible:ring-3 focus-visible:ring-blue-500/40 active:translate-y-px motion-safe:hover:-translate-y-0.5 sm:w-auto dark:border-blue-400/30 dark:bg-blue-500/10 dark:text-blue-300 dark:hover:bg-blue-500/20"
            variant="outline"
            render={
              isAuthenticated ? (
                <Link to="/dashboard" />
              ) : (
                <Link to="/sign-in" />
              )
            }
          >
            <Rocket className="mr-2 size-[18px]" />
            {t("Get Started")}
          </Button>
          <Button
            className="h-[54px] w-full rounded-xl border-slate-200 bg-white/70 px-7 text-[15px] font-medium text-slate-700 shadow-[0_8px_20px_-16px_rgba(15,23,42,0.28)] backdrop-blur-sm transition-all duration-200 hover:border-slate-300 hover:bg-white hover:shadow-[0_12px_24px_-14px_rgba(15,23,42,0.3)] focus-visible:outline-none focus-visible:ring-3 focus-visible:ring-slate-500/30 active:translate-y-px motion-safe:hover:-translate-y-0.5 sm:w-auto dark:border-white/15 dark:bg-white/5 dark:text-slate-200"
            variant="outline"
            render={<Link to="/pricing" />}
          >
            <Tags className="mr-2 size-[18px]" />
            {t("View Pricing")}
          </Button>
        </AnimateInView>

        <AnimateInView
          className="mt-7 text-center text-xs text-slate-400 dark:text-slate-500"
          animation="fade-in"
          delay={260}
        >
          {t("Sign up with free trial credits — email, phone or WeChat login")}
        </AnimateInView>
      </div>
    </section>
  );
}
