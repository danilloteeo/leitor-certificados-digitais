import { useEffect, useRef, useState } from "react";
import { Eye, EyeOff, FileKey2, Loader2, RotateCcw } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { CertDetails } from "@/components/cert-details";
import { CertInfo, ipc } from "@/lib/format";
import { ReadPFX, SelectPFXFile } from "../../wailsjs/go/main/App";
import { OnFileDrop, OnFileDropOff } from "../../wailsjs/runtime/runtime";

const fileName = (p: string) => p.split(/[\\/]/).pop() ?? p;

export function PfxView() {
  const [path, setPath] = useState("");
  const [password, setPassword] = useState("");
  const [showPwd, setShowPwd] = useState(false);
  const [loading, setLoading] = useState(false);
  const [info, setInfo] = useState<CertInfo | null>(null);
  const pwdRef = useRef<HTMLInputElement>(null);

  function choose(p: string) {
    setPath(p);
    setInfo(null);
    setPassword("");
    setTimeout(() => pwdRef.current?.focus(), 0);
  }

  useEffect(() => {
    OnFileDrop((_x, _y, paths) => {
      const pfx = paths.find((p) => /\.(pfx|p12)$/i.test(p));
      if (pfx) choose(pfx);
    }, true);
    return () => OnFileDropOff();
  }, []);

  async function pick() {
    const p = await ipc(() => SelectPFXFile());
    if (p) choose(p);
  }

  async function read(e: React.FormEvent) {
    e.preventDefault();
    if (!path) return;
    setLoading(true);
    try {
      setInfo(await ipc(() => ReadPFX(path, password)));
    } catch {
      pwdRef.current?.select();
    } finally {
      setLoading(false);
    }
  }

  function reset() {
    setPath("");
    setPassword("");
    setInfo(null);
  }

  if (info) {
    return (
      <div className="mx-auto flex w-full max-w-2xl flex-col gap-4 p-6">
        <div className="flex items-center justify-between gap-4">
          <p className="truncate text-sm text-muted-foreground" title={path}>
            <FileKey2 className="mr-1.5 inline size-4 align-[-3px]" />
            {fileName(path)}
          </p>
          <Button variant="outline" size="sm" onClick={reset}>
            <RotateCcw /> Ler outro certificado
          </Button>
        </div>
        <CertDetails c={info} />
      </div>
    );
  }

  return (
    <div className="grid h-full place-items-center p-6">
      <form onSubmit={read} className="flex w-full max-w-md flex-col gap-5">
        <button
          type="button"
          onClick={pick}
          style={{ "--wails-drop-target": "drop" } as React.CSSProperties}
          className="flex flex-col items-center gap-3 rounded-xl border-2 border-dashed px-6 py-10 text-center transition-colors hover:border-primary/60 hover:bg-muted/40 focus-visible:ring-3 focus-visible:ring-ring/50 focus-visible:outline-none"
        >
          <div className="grid size-12 place-items-center rounded-full bg-primary/10 text-primary">
            <FileKey2 className="size-6" />
          </div>
          {path ? (
            <>
              <span className="max-w-full truncate font-medium" title={path}>{fileName(path)}</span>
              <span className="text-xs text-muted-foreground">Clique para trocar o arquivo</span>
            </>
          ) : (
            <>
              <span className="font-medium">Arraste o certificado A1 aqui</span>
              <span className="text-xs text-muted-foreground">
                ou <span className="text-primary underline-offset-2 hover:underline">procure no computador</span> — .pfx ou .p12
              </span>
            </>
          )}
        </button>

        <div className="flex flex-col gap-1.5">
          <Label htmlFor="pwd">Senha do certificado</Label>
          <div className="relative">
            <Input
              id="pwd"
              ref={pwdRef}
              type={showPwd ? "text" : "password"}
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              disabled={!path}
              autoComplete="off"
              className="h-9 pr-9"
            />
            <button
              type="button"
              onClick={() => setShowPwd((v) => !v)}
              aria-label={showPwd ? "Ocultar senha" : "Mostrar senha"}
              className="absolute inset-y-0 right-0 grid w-9 place-items-center text-muted-foreground hover:text-foreground"
            >
              {showPwd ? <EyeOff className="size-4" /> : <Eye className="size-4" />}
            </button>
          </div>
        </div>

        <div className="flex justify-end">
          <Button type="submit" className="h-9" disabled={!path || loading}>
            {loading && <Loader2 className="animate-spin" />}
            Ler certificado
          </Button>
        </div>
      </form>
    </div>
  );
}
