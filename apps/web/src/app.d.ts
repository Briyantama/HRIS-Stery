declare global {
	namespace App {
		interface Locals {
			token?: string;
			user?: {
				user_id: string;
				tenant_id: string;
				email: string;
				roles: string[];
			};
		}
	}
}

export {};
