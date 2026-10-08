import { useEffect, useState } from "react";
import { Eye, EyeOff, Loader2 } from "lucide-react";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { CertInfo, ipc } from "@/lib/format";
import { ExportCertificate } from "../../wailsjs/go/main/App";

export function ExportDialog({ cert, open, onOpenChange }: { cert: CertInfo; open: boolean; onOpenChange: (o: boolean) => void }) {
  const [pwd, setPwd] = useState("");
  const [confirm, setConfirm] = useState("");
  const [show, setShow] = useState(false);
  const [pending, setPending] = useState(false);

  useEffect(() => {
    if (open) {
      setPwd("");
      setConfirm("");
      setShow(false);
    }
  }, [open]);

  const curta = pwd.length > 0 && pwd.length < 4;
  const diferente = confirm.length > 0 && confirm !== pwd;
  const valido = pwd.length >= 4 && confirm === pwd;

  async function submit(e: React.FormEvent) {
    e.preventDefault();
    if (!valido || pending) return;
    setPending(true);
    try {
      const path = await ipc(() => ExportCertificate(cert.thumbprintSha1, pwd));
      if (path) {
        toast.success("Certificado exportado", { description: path });
        onOpenChange(false);
      }
    } catch {
      /* erro já exibido */
    } finally {
      setPending(false);
    }
  }

  return (
    <Dialog open={open} onOpenChange={(o) => !pending && onOpenChange(o)}>
      <DialogContent className="sm:max-w-md">
        <form onSubmit={submit} className="grid gap-4">
          <DialogHeader>
            <DialogTitle>Exportar como PFX</DialogTitle>
            <DialogDescription>
              Gera um arquivo .pfx de <b className="font-medium text-foreground">{cert.titular || cert.commonName}</b> com
              a chave privada, protegido pela nova senha.
            </DialogDescription>
          </DialogHeader>

          {cert.nivel === "A3" && (
            <p className="rounded-md bg-warning/15 px-3 py-2 text-xs text-warning">
              Certificados A3 ficam no token/cartão e normalmente não permitem exportar a chave.
            </p>
          )}

          <div className="grid gap-1.5">
            <Label htmlFor="exp-pwd">Nova senha</Label>
            <div className="relative">
              <Input
                id="exp-pwd"
                autoFocus
                type={show ? "text" : "password"}
                value={pwd}
                onChange={(e) => setPwd(e.target.value)}
                autoComplete="new-password"
                aria-invalid={curta || undefined}
                className="h-9 pr-9"
              />
              <button
                type="button"
                onClick={() => setShow((v) => !v)}
                aria-label={show ? "Ocultar senha" : "Mostrar senha"}
                className="absolute inset-y-0 right-0 grid w-9 place-items-center text-muted-foreground hover:text-foreground"
              >
                {show ? <EyeOff className="size-4" /> : <Eye className="size-4" />}
              </button>
            </div>
            {curta && <p className="text-xs text-destructive">Use pelo menos 4 caracteres.</p>}
          </div>

          <div className="grid gap-1.5">
            <Label htmlFor="exp-confirm">Confirmar senha</Label>
            <Input
              id="exp-confirm"
              type={show ? "text" : "password"}
              value={confirm}
              onChange={(e) => setConfirm(e.target.value)}
              autoComplete="new-password"
              aria-invalid={diferente || undefined}
              className="h-9"
            />
            {diferente && <p className="text-xs text-destructive">As senhas não conferem.</p>}
          </div>

          <DialogFooter>
            <Button type="button" variant="outline" onClick={() => onOpenChange(false)} disabled={pending}>
              Cancelar
            </Button>
            <Button type="submit" disabled={!valido || pending}>
              {pending && <Loader2 className="animate-spin" />}
              Exportar…
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
