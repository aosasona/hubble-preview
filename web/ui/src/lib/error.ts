import { toast } from "sonner";
import { ProcedureCallError } from "./server/bindings";
import type { FieldValues, Path, UseFormSetError } from "react-hook-form";

type ExtractedError =
	| { type: "validation"; message: string; errors?: Record<string, string[]> }
	| { type: "authz"; message: string }
	| { type: "generic"; message: string }
	| { type: "app"; message: string };

type ValidationErrorDetails = {
	type: "validation-error";
	errors: Record<string, string[]>;
};

type AuthzErrorDetails = {
	type: "authz-error";
	message: string;
};

type ErrorDetails = ValidationErrorDetails | AuthzErrorDetails | string;

export class AppError extends Error {
	constructor(message: string) {
		super(message);
		this.name = "AppError";
	}
}

export function toastError(error: Error) {
	const extracted = extractError(error);
	if (!extracted) return;

	toast.error(extracted.message);
}

// Extracts the appropriate error shape from a ProcedureCallError
export function extractError(error: unknown): ExtractedError | null {
	if (!error) {
		return null;
	}

	if (error instanceof AppError) {
		return { type: "app", message: error.message };
	}

	if (!(error instanceof ProcedureCallError)) {
		return {
			type: "generic",
			message:
				"An error occurred, we are working on it, please try again later!",
		};
	}

	const details = error.details as ErrorDetails;

	if (typeof details === "string") {
		let msg = details;

		if (
			details.toLowerCase() === "load failed" ||
			details.startsWith("Failed to call procedure")
		) {
			msg = "Unable to connect to the server, please try again later!";
		}

		return { type: "generic", message: msg };
	}

	switch (details.type) {
		case "validation-error":
			return {
				type: "validation",
				message: "One or more fields are invalid",
				errors: details.errors,
			};

		case "authz-error":
			return { type: "authz", message: details.message };
	}
}

/**
 * Returns a function that sets the error message on the form handler for a mutation.
 *
 * It shows a toast with the error message and sets the error message on the form handler when it's a validation error.
 *
 * @param setError The function to set the error message on the form handler.
 *
 * ```ts
 * const mutation = useMutation({
 *  onMutate: async (data: Record<string, string>) => {
 *      // do something
 *  },
 *  onError: onMutationError(setError),
 * });
 *```
 */
export function onMutationError<T extends object = FieldValues>(
	setError?: UseFormSetError<T>,
): (error: unknown) => void {
	return (error: unknown) => {
		const extracted = extractError(error);
		if (!extracted) return;

		if (extracted.type === "validation") {
			if (!setError) {
				toast.error(extracted.message);
				return;
			}

			for (const [key, value] of Object.entries(extracted.errors ?? {})) {
				setError(key as Path<T>, { message: value?.join(", ") });
			}
			toast.error(extracted.message);
			return;
		}

		toast.error(extracted.message);
	};
}
