import { toast } from "sonner";
import { cert } from "../../wailsjs/go/models";
import { ClipboardSetText } from "../../wailsjs/runtime/runtime";

export type CertInfo = cert.Info;

/** Executa uma chamada ao Go e mostra o erro em um toast. */
export async function ipc<T>(call: () => Promise<T>): Promise<T> {
  try {
    return await call();
  } catch (e) {
    toast.error(String(e));
    throw e;
  }
}

export function formatDate(iso: string, withTime = false): string {
  if (!iso) return "—";
  const d = new Date(iso);
  return withTime
    ? d.toLocaleString("pt-BR", { dateStyle: "short", timeStyle: "short" })
    : d.toLocaleDateString("pt-BR");
}

export function documento(c: CertInfo): string {
  return c.cnpj || c.cpf || "";
}

export const STATUS: Record<string, { label: string; className: string; dot: string }> = {
  valido: { label: "Válido", className: "bg-success/12 text-success", dot: "bg-success" },
  vencendo: { label: "Vencendo", className: "bg-warning/15 text-warning", dot: "bg-warning" },
  vencido: { label: "Vencido", className: "bg-destructive/12 text-destructive", dot: "bg-destructive" },
  pendente: { label: "Ainda não vigente", className: "bg-muted text-muted-foreground", dot: "bg-muted-foreground" },
};

export function prazo(c: CertInfo): string {
  const d = c.diasRestantes;
  if (c.status === "pendente") return "ainda não está em vigor";
  if (d < 0) return `vencido há ${-d} ${-d === 1 ? "dia" : "dias"}`;
  if (d === 0) return "vence hoje";
  return `vence em ${d} ${d === 1 ? "dia" : "dias"}`;
}

export async function copiar(texto: string, rotulo: string) {
  try {
    const ok = await ClipboardSetText(texto);
    if (!ok) throw new Error();
    toast.success(`${rotulo} copiado`);
  } catch {
    toast.error("Não foi possível copiar");
  }
}
