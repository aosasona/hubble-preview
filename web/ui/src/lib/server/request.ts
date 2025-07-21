import { ProcedureCallError, type ServerResponse } from "./bindings";

export const ServerAddr = import.meta.env.DEV ? "http://localhost:3288" : "";

export async function makeApiCall<T>(
	route: string,
	options: RequestInit,
): Promise<T> {
	try {
		const path = route.startsWith("/") ? route : `/${route}`;
		const response = await fetch(`${ServerAddr}/api/v1${path}`, options);

		if (!response.ok) {
			let err: unknown = "An unknown error occurred";

			// Attempt to parse the response body as JSON to extract the error message
			try {
				const data = (await response.json()) as ServerResponse<T>;
				if (!!data && data?.error) {
					err = data?.error;
				}
				// eslint-disable-next-line @typescript-eslint/no-unused-vars
			} catch (_) {
				/* Ignore errors here and just throw anyway */
			}
			throw new ProcedureCallError(err, path);
		}

		const data = (await response.json()) as ServerResponse<T>;
		if (!data.ok) {
			throw new ProcedureCallError(
				data?.error || "An unknown error occurred",
				path,
			);
		}

		return data?.data as T;
	} catch (e) {
		if (e instanceof ProcedureCallError) {
			throw e;
		}

		const message = Object.prototype.hasOwnProperty.call(e, "message")
			? (e as { message: unknown }).message
			: "An unknown error occurred";
		throw new ProcedureCallError(message, route, e as Error);
	}
}
