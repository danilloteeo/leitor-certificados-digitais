export namespace cert {
	
	export class Info {
	    titular: string;
	    commonName: string;
	    tipo: string;
	    nivel: string;
	    cpf: string;
	    cnpj: string;
	    responsavel: string;
	    responsavelCpf: string;
	    dataNascimento: string;
	    email: string;
	    numeroSerie: string;
	    numeroSerieDec: string;
	    emissor: string;
	    emissorCompleto: string;
	    subjectCompleto: string;
	    validoDe: string;
	    validoAte: string;
	    diasRestantes: number;
	    status: string;
	    thumbprintSha1: string;
	    thumbprintSha256: string;
	    algoritmoAssinatura: string;
	    algoritmoChave: string;
	    possuiChavePrivada: boolean;
	    origem: string;
	
	    static createFrom(source: any = {}) {
	        return new Info(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.titular = source["titular"];
	        this.commonName = source["commonName"];
	        this.tipo = source["tipo"];
	        this.nivel = source["nivel"];
	        this.cpf = source["cpf"];
	        this.cnpj = source["cnpj"];
	        this.responsavel = source["responsavel"];
	        this.responsavelCpf = source["responsavelCpf"];
	        this.dataNascimento = source["dataNascimento"];
	        this.email = source["email"];
	        this.numeroSerie = source["numeroSerie"];
	        this.numeroSerieDec = source["numeroSerieDec"];
	        this.emissor = source["emissor"];
	        this.emissorCompleto = source["emissorCompleto"];
	        this.subjectCompleto = source["subjectCompleto"];
	        this.validoDe = source["validoDe"];
	        this.validoAte = source["validoAte"];
	        this.diasRestantes = source["diasRestantes"];
	        this.status = source["status"];
	        this.thumbprintSha1 = source["thumbprintSha1"];
	        this.thumbprintSha256 = source["thumbprintSha256"];
	        this.algoritmoAssinatura = source["algoritmoAssinatura"];
	        this.algoritmoChave = source["algoritmoChave"];
	        this.possuiChavePrivada = source["possuiChavePrivada"];
	        this.origem = source["origem"];
	    }
	}

}

