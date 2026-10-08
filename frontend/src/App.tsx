import { useState } from "react";
import { Moon, Sun } from "lucide-react";
import { Toaster } from "@/components/ui/sonner";
import { ThemeProvider, useTheme } from "@/components/theme-provider";
import { PfxView } from "@/views/pfx-view";
import { StoreView } from "@/views/store-view";
import { cn } from "@/lib/utils";

const SCREENS = [
  { id: "store", label: "Instalados" },
  { id: "pfx", label: "Abrir arquivo PFX" },
] as const;

type ScreenId = (typeof SCREENS)[number]["id"];

function Shell() {
  const [screen, setScreen] = useState<ScreenId>("store");
  const { resolved, setTheme } = useTheme();

  return (
    <div className="flex h-full w-full flex-col bg-background text-foreground">
      <header className="flex h-12 shrink-0 items-center gap-6 border-b px-4">
        <span className="font-semibold">Certificados</span>
        <nav className="flex h-full gap-4">
          {SCREENS.map((s) => (
            <button
              key={s.id}
              onClick={() => setScreen(s.id)}
              className={cn(
                "-mb-px border-b-2 px-1 text-sm transition-colors",
                screen === s.id
                  ? "border-primary font-medium text-foreground"
                  : "border-transparent text-muted-foreground hover:text-foreground",
              )}
            >
              {s.label}
            </button>
          ))}
        </nav>
        <button
          onClick={() => setTheme(resolved === "dark" ? "light" : "dark")}
          aria-label={resolved === "dark" ? "Tema claro" : "Tema escuro"}
          className="ml-auto grid size-8 place-items-center rounded-md text-muted-foreground hover:bg-muted hover:text-foreground"
        >
          {resolved === "dark" ? <Sun className="size-4" /> : <Moon className="size-4" />}
        </button>
      </header>
      <main className="min-h-0 flex-1 overflow-y-auto">{screen === "store" ? <StoreView /> : <PfxView />}</main>
    </div>
  );
}

export default function App() {
  return (
    <ThemeProvider>
      <Shell />
      <Toaster position="bottom-right" richColors />
    </ThemeProvider>
  );
}
