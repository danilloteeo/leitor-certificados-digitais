import { useCallback, useEffect, useMemo, useState } from "react";
import { Download, RefreshCw, Search, ShieldOff, Trash2, X } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Checkbox } from "@/components/ui/checkbox";
import { DeleteDialog } from "@/components/delete-dialog";
import { ExportDialog } from "@/components/export-dialog";
import { Input } from "@/components/ui/input";
import { Skeleton } from "@/components/ui/skeleton";
import { CertDetails } from "@/components/cert-details";
import { CertInfo, documento, formatDate, ipc, STATUS } from "@/lib/format";
import { cn } from "@/lib/utils";
import { ListStoreCertificates } from "../../wailsjs/go/main/App";

const FILTROS = [
  { id: "todos", label: "Todos" },
  { id: "valido", label: "Válidos" },
  { id: "vencendo", label: "Vencendo" },
  { id: "vencido", label: "Vencidos" },
] as const;

type Filtro = (typeof FILTROS)[number]["id"];

const norm = (s: string) => s.normalize("NFD").replace(/[\u0300-\u036f]/g, "").toLowerCase();
const digits = (s: string) => s.replace(/\D/g, "");

export function StoreView() {
  const [certs, setCerts] = useState<CertInfo[] | null>(null);
  const [selected, setSelected] = useState<string | null>(null);
  const [query, setQuery] = useState("");
  const [filtro, setFiltro] = useState<Filtro>("todos");
  const [marked, setMarked] = useState<Set<string>>(new Set());
  const [exporting, setExporting] = useState<CertInfo | null>(null);
  const [deleting, setDeleting] = useState<CertInfo[] | null>(null);

  const load = useCallback(async () => {
    setCerts(null);
    try {
      const list = (await ipc(() => ListStoreCertificates())) ?? [];
      setCerts(list);
      // mantém marcados só os que ainda existem
      setMarked((prev) => new Set(list.filter((c) => prev.has(c.thumbprintSha1)).map((c) => c.thumbprintSha1)));
    } catch {
      setCerts([]);
    }
  }, []);

  useEffect(() => {
    load();
  }, [load]);

  const counts = useMemo(() => {
    const c: Record<string, number> = { todos: certs?.length ?? 0 };
    certs?.forEach((x) => (c[x.status] = (c[x.status] ?? 0) + 1));
    return c;
  }, [certs]);

  const visible = useMemo(() => {
    if (!certs) return [];
    const q = norm(query.trim());
    const qd = digits(query);
    return certs.filter((c) => {
      if (filtro !== "todos" && c.status !== filtro) return false;
      if (!q) return true;
      return (
        norm(c.titular + " " + c.commonName + " " + c.emissor).includes(q) ||
        (qd.length >= 3 && digits(documento(c)).includes(qd)) ||
        c.numeroSerie.toLowerCase().includes(q)
      );
    });
  }, [certs, query, filtro]);

  const markedCerts = useMemo(() => (certs ?? []).filter((c) => marked.has(c.thumbprintSha1)), [certs, marked]);
  const allVisibleMarked = visible.length > 0 && visible.every((c) => marked.has(c.thumbprintSha1));
  const someVisibleMarked = visible.some((c) => marked.has(c.thumbprintSha1));

  function toggle(thumb: string, on: boolean) {
    setMarked((prev) => {
      const next = new Set(prev);
      if (on) next.add(thumb);
      else next.delete(thumb);
      return next;
    });
  }

  function toggleAllVisible(on: boolean) {
    setMarked((prev) => {
      const next = new Set(prev);
      visible.forEach((c) => (on ? next.add(c.thumbprintSha1) : next.delete(c.thumbprintSha1)));
      return next;
    });
  }

  const current = visible.find((c) => c.thumbprintSha1 === selected) ?? visible[0] ?? null;

  function onKeyDown(e: React.KeyboardEvent) {
    if (e.key !== "ArrowDown" && e.key !== "ArrowUp") return;
    e.preventDefault();
    const i = current ? visible.indexOf(current) : -1;
    const next = visible[Math.max(0, Math.min(visible.length - 1, i + (e.key === "ArrowDown" ? 1 : -1)))];
    if (next) {
      setSelected(next.thumbprintSha1);
      document.getElementById(`cert-${next.thumbprintSha1}`)?.scrollIntoView({ block: "nearest" });
    }
  }

  return (
    <div className="flex h-full min-h-0">
      <aside className="flex w-[360px] shrink-0 flex-col border-r">
        <div className="flex flex-col gap-3 border-b p-3">
          <div className="flex gap-2">
            <div className="relative flex-1">
              <Search className="pointer-events-none absolute top-1/2 left-2.5 size-4 -translate-y-1/2 text-muted-foreground" />
              <Input
                value={query}
                onChange={(e) => setQuery(e.target.value)}
                onKeyDown={onKeyDown}
                placeholder="Nome, CPF, CNPJ ou série"
                className="h-9 pl-8"
              />
            </div>
            <Button variant="outline" size="icon" className="size-9" onClick={load} aria-label="Recarregar">
              <RefreshCw />
            </Button>
          </div>
          <div className="flex items-center gap-1">
            <Checkbox
              checked={allVisibleMarked}
              indeterminate={someVisibleMarked && !allVisibleMarked}
              onCheckedChange={(v) => toggleAllVisible(!!v)}
              disabled={visible.length === 0}
              aria-label="Selecionar todos os exibidos"
              className="mr-1.5 ml-1"
            />
            {FILTROS.map((f) => (
              <button
                key={f.id}
                onClick={() => setFiltro(f.id)}
                className={cn(
                  "h-7 rounded-md px-2 text-xs font-medium transition-colors",
                  filtro === f.id ? "bg-secondary text-foreground" : "text-muted-foreground hover:bg-muted/60",
                )}
              >
                {f.label}
                <span className="ml-1 tabular-nums opacity-60">{counts[f.id] ?? 0}</span>
              </button>
            ))}
          </div>
        </div>

        {marked.size > 0 && (
          <div className="flex h-11 items-center gap-2 border-b bg-muted/40 px-3 text-sm">
            <span className="flex-1 font-medium tabular-nums">
              {marked.size} {marked.size === 1 ? "selecionado" : "selecionados"}
            </span>
            <Button size="sm" variant="destructive" onClick={() => setDeleting(markedCerts)}>
              <Trash2 /> Remover
            </Button>
            <Button size="icon-sm" variant="ghost" onClick={() => setMarked(new Set())} aria-label="Limpar seleção">
              <X />
            </Button>
          </div>
        )}

        <div className="min-h-0 flex-1 overflow-y-auto" onKeyDown={onKeyDown}>
          {certs === null ? (
            <ListSkeleton />
          ) : visible.length === 0 ? (
            <p className="p-6 text-center text-sm text-muted-foreground">
              {certs.length === 0 ? "Nenhum certificado instalado." : "Nada encontrado com esse filtro."}
            </p>
          ) : (
            <ul className="divide-y">
              {visible.map((c) => {
                const active = c === current;
                return (
                  <li
                    key={c.thumbprintSha1}
                    className={cn("relative flex items-center transition-colors", active ? "bg-primary/8" : "hover:bg-muted/50")}
                  >
                    {active && <span className="absolute inset-y-0 left-0 w-0.5 bg-primary" />}
                    <Checkbox
                      checked={marked.has(c.thumbprintSha1)}
                      onCheckedChange={(v) => toggle(c.thumbprintSha1, !!v)}
                      aria-label={`Selecionar ${c.titular || c.commonName}`}
                      className="ml-4"
                    />
                    <button
                      id={`cert-${c.thumbprintSha1}`}
                      onClick={() => setSelected(c.thumbprintSha1)}
                      className="flex min-w-0 flex-1 flex-col gap-1 py-2.5 pr-4 pl-3 text-left focus-visible:bg-muted focus-visible:outline-none"
                    >
                      <span className="truncate font-medium">{c.titular || c.commonName}</span>
                      <span className="flex items-center justify-between gap-2 text-xs text-muted-foreground">
                        <span className="truncate font-mono">{documento(c) || c.tipo}</span>
                        <span
                          className="flex shrink-0 items-center gap-1.5 tabular-nums"
                          title={STATUS[c.status]?.label}
                        >
                          <span className={cn("size-1.5 rounded-full", STATUS[c.status]?.dot)} />
                          {formatDate(c.validoAte)}
                        </span>
                      </span>
                    </button>
                  </li>
                );
              })}
            </ul>
          )}
        </div>
      </aside>

      <section className="min-w-0 flex-1 overflow-y-auto">
        {current ? (
          <div className="mx-auto max-w-2xl p-6">
            <CertDetails
              c={current}
              actions={
                <>
                  {current.possuiChavePrivada && (
                    <Button variant="outline" size="sm" onClick={() => setExporting(current)}>
                      <Download /> Exportar PFX
                    </Button>
                  )}
                  <Button
                    variant="outline"
                    size="sm"
                    className="text-destructive hover:text-destructive"
                    onClick={() => setDeleting([current])}
                  >
                    <Trash2 /> Remover
                  </Button>
                </>
              }
            />
          </div>
        ) : certs === null ? (
          <div className="flex flex-col gap-4 p-6">
            <Skeleton className="h-11 w-2/3" />
            <Skeleton className="h-4 w-1/3" />
            <Skeleton className="h-40" />
          </div>
        ) : (
          <div className="grid h-full place-items-center p-6 text-center">
            <div className="flex flex-col items-center gap-2 text-muted-foreground">
              <ShieldOff className="size-8" />
              <p className="text-sm">
                Nenhum certificado no repositório pessoal do Windows.
                <br />
                Se usa token A3, conecte-o e clique em recarregar.
              </p>
              <Button variant="outline" size="sm" className="mt-2" onClick={load}>
                <RefreshCw /> Recarregar
              </Button>
            </div>
          </div>
        )}
      </section>

      {exporting && (
        <ExportDialog cert={exporting} open={!!exporting} onOpenChange={(o) => !o && setExporting(null)} />
      )}
      {deleting && (
        <DeleteDialog
          certs={deleting}
          open={!!deleting}
          onOpenChange={(o) => !o && setDeleting(null)}
          onDeleted={load}
        />
      )}
    </div>
  );
}

function ListSkeleton() {
  return (
    <div className="divide-y">
      {Array.from({ length: 8 }).map((_, i) => (
        <div key={i} className="flex flex-col gap-2 px-4 py-3">
          <Skeleton className="h-4 w-3/4" />
          <Skeleton className="h-3 w-1/2" />
        </div>
      ))}
    </div>
  );
}
