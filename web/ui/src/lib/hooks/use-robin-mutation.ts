import { useMutation, useQueryClient } from "@tanstack/react-query";
import type { FieldValues, UseFormSetError } from "react-hook-form";
import type { MutationPayload, MutationResult, ProcedureKey } from "../server";
import client from "../server";
import { onMutationError } from "../error";
import { toast } from "sonner";
import { useCallback } from "react";

type AsyncMode = "sync" | "async";

type MutationOptions<
	Key extends ProcedureKey<"mutation">,
	FV extends object = FieldValues,
	Mode extends AsyncMode = "sync",
> = {
	/** The mutation key to be used */
	mutationKey?: Array<string | object | number>;

	/** This will be used to call the mutation asynchronously */
	mode?: Mode;

	/** This will be called before the mutation is called */
	beforeCall?: (payload: MutationPayload<Key>) => void;

	/**
	 * This will be called when the mutation is successful
	 *
	 * If this is a string, it will be shown as a toast message
	 * If this is a function, it will be called with the mutation result as an argument
	 */
	onSuccess?: ((data: MutationResult<Key>) => void) | string;

	/** This will be called when the mutation fails */
	onError?: (error: Error) => void;

	/** This will be called when the mutation is complete regardless of the error or success state */
	onComplete?: () => void;

	/** This will be used to automatically set the form validation errors if present */
	// setFormError?: UseFormSetError<FV>;
	setFormError?: UseFormSetError<FV>;

	/** The query keys to be invalidated after the mutation is **SUCCESSFUL** */
	invalidates?:
		| Array<string | object | number>
		| Array<Array<string | object | number>>
		| Array<ProcedureKey<"query">>
		| Array<ProcedureKey<"mutation">>;

	/** The number of times to retry the mutation if it fails */
	retry?: boolean | number;
};

type State = {
	/** This is true if the mutation was successful */
	isSuccess: boolean;

	/** This is true if the mutation failed */
	isError: boolean;

	/** This is true if the mutation is in progress */
	isMutating: boolean;

	/** This is true if the mutation has not been called yet */
	isIdle: boolean;
};

type MutationHookResult<
	Key extends ProcedureKey<"mutation">,
	Mode extends AsyncMode,
> = {
	data?: MutationResult<Key>;
	error: Error | null;
	reset: () => void;
	call: Mode extends "async"
		? (payload: MutationPayload<Key>) => Promise<void>
		: (payload: MutationPayload<Key>) => void;
} & State;

export default function useRobinMutation<
	Key extends ProcedureKey<"mutation">,
	FV extends object = FieldValues,
	Mode extends AsyncMode = "async",
>(
	key: Key,
	options?: MutationOptions<Key, FV, Mode>,
): MutationHookResult<Key, Mode> {
	const queryClient = useQueryClient();

	const {
		data,
		error,
		mutate,
		mutateAsync,
		isSuccess,
		isError,
		isPending,
		isIdle,
		reset,
	} = useMutation({
		mutationKey: options?.mutationKey ?? key.split("."),
		mutationFn: (payload: MutationPayload<Key>) => {
			return client.call("mutation", { name: key, payload });
		},
		onSuccess: (result) => {
			if (options?.invalidates) {
				for (const key of options.invalidates) {
					const queryKey = Array.isArray(key)
						? key
						: typeof key === "string"
							? key.split(".")
							: [key];
					queryClient.invalidateQueries({ queryKey });
				}
			}

			if (options?.onSuccess) {
				if (typeof options.onSuccess === "string") {
					toast.success(options.onSuccess);
					return;
				}

				options.onSuccess(result);
				return;
			}

			if (typeof result === "object" && "message" in result) {
				toast.success(result.message);
			}
		},
		onError: (error) => {
			if (options?.onError) {
				options.onError(error);
				return;
			}

			onMutationError(options?.setFormError)(error);
		},
		onSettled: options?.onComplete,
		retry: options?.retry ?? 1,
	});

	const asyncCall = useCallback(
		async (payload: MutationPayload<Key>) => {
			if (options?.beforeCall) {
				options.beforeCall(payload);
			}

			await mutateAsync(payload);
		},
		[mutateAsync, options],
	);

	const syncCall = useCallback(
		(payload: MutationPayload<Key>) => {
			if (options?.beforeCall) {
				options.beforeCall(payload);
			}

			mutate(payload);
		},
		[mutate, options],
	);

	return {
		data,
		error,
		call: (options?.mode === undefined || options?.mode === "async"
			? asyncCall
			: syncCall) as MutationHookResult<Key, Mode>["call"],
		isSuccess,
		isError,
		isMutating: isPending,
		isIdle,
		reset,
	};
}
