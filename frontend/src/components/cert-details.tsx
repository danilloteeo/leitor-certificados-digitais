import { useState } from "react";
import { ChevronDown, Copy } from "lucide-react";
import { cn } from "@/lib/utils";
import { CertInfo, copiar, formatDate, prazo, STATUS } from "@/lib/format";

export function StatusBadge({ status }: { status: string }) {
  const s = STATUS[status] ?? STATUS.pendente;
  return (
    <span className={cn("inline-flex h-5 items-center gap-1.5 rounded-full px-2 text-xs font-medium", s.className)}>
      <span className={cn("size-1.5 rounded-full", s.dot)} />
      {s.label}
    </span>
  );
}

function Row({ label, value, mono, copy, children }: {
  label: string;
  value?: string;
  mono?: boolean;
  copy?: boolean;
  children?: React.ReactNode;
}) {
  if (!value && !children) return null;
  return (
    <div className="group grid grid-cols-[160px_1fr] items-baseline gap-4 py-2.5">
      <dt className="text-muted-foreground">{label}</dt>
      <dd className="flex min-w-0 items-baseline gap-1.5">
        <span className={cn("min-w-0 break-all", mono && "font-mono text-[13px]")}>{value}</span>
        {children}
        {copy && value && (
          <button
            type="button"
            onClick={() => copiar(value, label)}
            aria-label={`Copiar ${label}`}
            className="shrink-0 self-center rounded p-0.5 text-muted-foreground opacity-0 transition-opacity group-hover:opacity-100 hover:text-foreground focus-visible:opacity-100"
          >
            <Copy className="size-3.5" />
          </button>
        )}
      </dd>
    </div>
  );
}

export function CertDetails({ c, actions }: { c: CertInfo; actions?: React.ReactNode }) {
  const [more, setMore] = useState(false);
  const doc = c.cnpj || c.cpf;
  const prazoCor = STATUS[c.status]?.className.split(" ").find((x) => x.startsWith("text-"));

  return (
    <div className="flex flex-col gap-5">
      <header className="flex items-start justify-between gap-4">
        <div className="min-w-0">
          <h2 className="text-lg leading-tight font-semibold break-words">{c.titular || c.commonName}</h2>
          <p className="mt-1 flex flex-wrap items-center gap-2 text-sm text-muted-foreground">
            {[c.tipo, c.nivel].filter(Boolean).join(" · ")}
            <StatusBadge status={c.status} />
          </p>
        </div>
        {actions && <div className="flex shrink-0 gap-2">{actions}</div>}
      </header>

      <dl className="divide-y border-y">
        <Row label={c.cnpj ? "CNPJ" : "CPF"} value={doc || "—"} mono copy={!!doc} />
        <Row label="Número de série" value={c.numeroSerie} mono copy />
        <Row label="Vencimento" value={formatDate(c.validoAte)}>
          <span className={cn("text-xs", prazoCor)}>({prazo(c)})</span>
        </Row>
        <Row label="Emitido por" value={c.emissor} />
        <Row label="Responsável" value={c.responsavel} />
      </dl>

      <div>
        <button
          type="button"
          onClick={() => setMore((v) => !v)}
          className="inline-flex items-center gap-1 text-sm text-muted-foreground hover:text-foreground"
        >
          <ChevronDown className={cn("size-4 transition-transform", more && "rotate-180")} />
          {more ? "Menos detalhes" : "Mais detalhes"}
        </button>

        {more && (
          <dl className="mt-2 divide-y border-y">
            <Row label="CPF do responsável" value={c.responsavelCpf} mono copy />
            <Row label="Data de nascimento" value={c.dataNascimento} />
            <Row label="E-mail" value={c.email} copy />
            <Row label="Válido desde" value={formatDate(c.validoDe, true)} />
            <Row label="Válido até" value={formatDate(c.validoAte, true)} />
            <Row label="Série (decimal)" value={c.numeroSerieDec} mono copy />
            <Row label="Chave privada" value={c.possuiChavePrivada ? "Sim" : "Não"} />
            <Row label="Algoritmo" value={`${c.algoritmoChave} · ${c.algoritmoAssinatura}`} />
            <Row label="Thumbprint SHA-1" value={c.thumbprintSha1} mono copy />
            <Row label="Thumbprint SHA-256" value={c.thumbprintSha256} mono copy />
            <Row label="Subject" value={c.subjectCompleto} mono copy />
            <Row label="Emissor completo" value={c.emissorCompleto} mono copy />
            <Row label="Origem" value={c.origem} />
          </dl>
        )}
      </div>
    </div>
  );
}
