import { createContext, useContext, useEffect, useState } from "react";
import { WindowSetDarkTheme, WindowSetLightTheme } from "../../wailsjs/runtime/runtime";

type Theme = "light" | "dark" | "system";

const ThemeCtx = createContext<{ theme: Theme; resolved: "light" | "dark"; setTheme: (t: Theme) => void }>({
  theme: "system",
  resolved: "light",
  setTheme: () => {},
});

function readStored(): Theme {
  try {
    return (localStorage.getItem("theme") as Theme) || "system";
  } catch {
    return "system";
  }
}

export function ThemeProvider({ children }: { children: React.ReactNode }) {
  const [theme, setTheme] = useState<Theme>(readStored);
  const [sysDark, setSysDark] = useState(() => window.matchMedia("(prefers-color-scheme: dark)").matches);

  useEffect(() => {
    const mq = window.matchMedia("(prefers-color-scheme: dark)");
    const onChange = (e: MediaQueryListEvent) => setSysDark(e.matches);
    mq.addEventListener("change", onChange);
    return () => mq.removeEventListener("change", onChange);
  }, []);

  const resolved = theme === "dark" || (theme === "system" && sysDark) ? "dark" : "light";

  useEffect(() => {
    document.documentElement.classList.toggle("dark", resolved === "dark");
    try {
      localStorage.setItem("theme", theme);
    } catch {
      /* preferência não persistida */
    }
    try {
      if (resolved === "dark") WindowSetDarkTheme();
      else WindowSetLightTheme();
    } catch {
      /* fora do Wails (ex.: vite no navegador) */
    }
  }, [theme, resolved]);

  return <ThemeCtx.Provider value={{ theme, resolved, setTheme }}>{children}</ThemeCtx.Provider>;
}

export const useTheme = () => useContext(ThemeCtx);
