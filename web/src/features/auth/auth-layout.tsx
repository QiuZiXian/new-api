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
import { BarChart3, KeyRound, ShieldCheck } from "lucide-react";
import { useTranslation } from "react-i18next";

import { PublicLayout } from "@/components/layout";
import { Skeleton } from "@/components/ui/skeleton";
import { useSystemConfig } from "@/hooks/use-system-config";

type AuthLayoutProps = {
  children: React.ReactNode;
};

function BrandFeature(props: {
  icon: React.ReactNode;
  title: string;
  desc: string;
}) {
  return (
    <div className="flex items-start gap-3">
      <span className="flex size-9 shrink-0 items-center justify-center rounded-lg bg-white/15 text-white backdrop-blur-xs">
        {props.icon}
      </span>
      <div className="min-w-0">
        <p className="text-sm font-semibold text-white">{props.title}</p>
        <p className="mt-0.5 text-xs leading-relaxed text-blue-100/85">
          {props.desc}
        </p>
      </div>
    </div>
  );
}

export function AuthLayout({ children }: AuthLayoutProps) {
  const { t } = useTranslation();
  const { systemName, loading } = useSystemConfig();

  return (
    <PublicLayout
      showMainContainer={false}
      headerProps={{ className: "text-white" }}
    >
      <div className="relative isolate grid min-h-svh lg:grid-cols-2">
        {/* Full-bleed background artwork, mirrored so the illustration sits
            on the left and the open space stays behind the form. */}
        <img
          aria-hidden
          src="/home/hero-slide-2.jpg"
          alt=""
          className="absolute inset-0 -z-20 h-full w-full scale-x-[-1] object-cover object-[65%_center]"
        />
        <div
          aria-hidden
          className="absolute inset-0 -z-10 bg-[linear-gradient(90deg,rgba(6,17,44,0.38)_0%,rgba(6,17,44,0.20)_45%,rgba(6,17,44,0.06)_100%)]"
        />

        {/* Left: brand panel (desktop only) */}
        <aside className="relative hidden lg:flex lg:flex-col">
          {/* Value proposition */}
          <div className="relative z-10 flex flex-1 flex-col justify-center gap-8 px-10 pt-24 pb-12">
            <div>
              <h2 className="text-3xl leading-snug font-bold tracking-tight text-white">
                {t("Welcome back")}
              </h2>
              <p className="mt-3 max-w-md text-sm leading-relaxed text-blue-100/90">
                {t(
                  "Sign in to manage your API keys, model usage, billing and enterprise AI access.",
                )}
              </p>
            </div>

            <div className="space-y-5">
              <BrandFeature
                icon={<KeyRound className="size-4" />}
                title={t("Unified model access")}
                desc={t(
                  "One key for text, image, video and multimodal models.",
                )}
              />
              <BrandFeature
                icon={<ShieldCheck className="size-4" />}
                title={t("Enterprise-grade access control")}
                desc={t("Quotas, expiry, model limits and IP allowlists.")}
              />
              <BrandFeature
                icon={<BarChart3 className="size-4" />}
                title={t("Usage & cost visibility")}
                desc={t(
                  "Track tokens, requests, costs and task history in real time.",
                )}
              />
            </div>
          </div>

          <p className="relative z-10 px-10 pb-8 text-xs text-blue-100/70">
            © {new Date().getFullYear()}{" "}
            {loading ? (
              <Skeleton className="inline-block h-3 w-24" />
            ) : (
              systemName
            )}
          </p>
        </aside>

        {/* Right: form area (top bar comes from PublicLayout) */}
        <main className="relative flex flex-col">
          <div className="container flex flex-1 items-center justify-center pt-24 pb-10 lg:pt-28">
            <div className="glass-1 relative mx-auto flex w-full flex-col justify-center space-y-2 border-border/50 px-6 py-8 shadow-[0_20px_60px_-25px_rgba(15,23,42,0.18)] sm:w-[440px] sm:rounded-2xl sm:p-8 dark:shadow-[0_20px_60px_-25px_rgba(0,0,0,0.6)]">
              {children}
            </div>
          </div>
        </main>
      </div>
    </PublicLayout>
  );
}
