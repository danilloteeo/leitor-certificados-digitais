import { useState } from "react";
import { Loader2 } from "lucide-react";
import { toast } from "sonner";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog";
import { CertInfo, documento, ipc } from "@/lib/format";
import { DeleteCertificates } from "../../wailsjs/go/main/App";

const MAX_LISTA = 5;

export function DeleteDialog({
  certs,
  open,
  onOpenChange,
  onDeleted,
}: {
  certs: CertInfo[];
  open: boolean;
  onOpenChange: (o: boolean) => void;
  onDeleted: () => void;
}) {
  const [pending, setPending] = useState(false);
  const comChave = certs.filter((c) => c.possuiChavePrivada).length;

  async function confirm() {
    setPending(true);
    try {
      const n = await ipc(() => DeleteCertificates(certs.map((c) => c.thumbprintSha1)));
      toast.success(n === 1 ? "Certificado removido" : `${n} certificados removidos`);
      onOpenChange(false);
    } catch {
      onOpenChange(false); // remoção parcial: o erro já foi exibido
    } finally {
      setPending(false);
      onDeleted();
    }
  }

  return (
    <AlertDialog open={open} onOpenChange={(o) => !pending && onOpenChange(o)}>
      <AlertDialogContent className="data-[size=default]:sm:max-w-md">
        <AlertDialogHeader>
          <AlertDialogTitle>
            {certs.length === 1 ? "Remover certificado?" : `Remover ${certs.length} certificados?`}
          </AlertDialogTitle>
          <AlertDialogDescription>
            Eles serão removidos do repositório pessoal do Windows. Esta ação não pode ser desfeita.
            {comChave > 0 && " Se ainda precisar deles, exporte como PFX antes."}
          </AlertDialogDescription>
        </AlertDialogHeader>

        <ul className="max-h-48 divide-y overflow-y-auto rounded-md border text-sm">
          {certs.slice(0, MAX_LISTA).map((c) => (
            <li key={c.thumbprintSha1} className="flex justify-between gap-3 px-3 py-2">
              <span className="truncate">{c.titular || c.commonName}</span>
              <span className="shrink-0 font-mono text-xs text-muted-foreground">{documento(c)}</span>
            </li>
          ))}
          {certs.length > MAX_LISTA && (
            <li className="px-3 py-2 text-xs text-muted-foreground">e mais {certs.length - MAX_LISTA}…</li>
          )}
        </ul>

        <AlertDialogFooter>
          <AlertDialogCancel disabled={pending}>Cancelar</AlertDialogCancel>
          <AlertDialogAction variant="destructive" onClick={confirm} disabled={pending}>
            {pending && <Loader2 className="animate-spin" />}
            Remover
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  );
}
