import {
	type QueryObserverResult,
	type RefetchOptions,
	useQuery,
} from "@tanstack/react-query";
import type { ProcedureKey, QueryPayload, QueryResult } from "../server";
import client from "../server";

type QueryOptions<Data> = {
	/** The query key to be used */
	queryKey?: Array<string | object | number>;

	/** The placeholder data to be used while the query is loading */
	placeholderData?: Data | (() => Data);

	/** If this is false, the query will not be executed until it is enabled */
	enabled?: boolean;

	/** The number of times the query should be retried */
	retry?: boolean | number | { count: number; delay: number };

	/** When the data is considered stale */
	staleTime?: number;
};

type State = {
	/** This is true if the query was successful */
	isSuccess: boolean;

	/** This is true if the query failed */
	isError: boolean;

	/** This is true if the query is in progress */
	isLoading: boolean;

	/** This is true if the query has not been called yet */
	isFetched: boolean;

	/** This is true if the query is using placeholder data */
	isPlaceholderData: boolean;
};

type QueryHookResult<Data> = {
	data?: Data;
	error: Error | null;
	refetch: (
		options: RefetchOptions,
	) => Promise<QueryObserverResult<Data, Error>>;
} & State;

export default function useRobinQuery<Key extends ProcedureKey<"query">>(
	key: Key,
	params?: QueryPayload<Key>,
	options?: QueryOptions<QueryResult<Key>>,
): QueryHookResult<QueryResult<Key>> {
	const opts: Record<string, unknown> = {};
	if (options?.placeholderData) {
		opts.placeholderData =
			typeof options.placeholderData === "function"
				? options.placeholderData
				: () => options.placeholderData;
	}

	if (options?.enabled !== undefined) {
		opts.enabled = options.enabled;
	}

	if (options?.retry !== undefined) {
		opts.retry = options.retry;
	}

	if (options?.staleTime !== undefined) {
		opts.staleTime = options.staleTime;
	}

	const { data, error, status, isPlaceholderData, isFetched, refetch } =
		useQuery({
			queryKey: options?.queryKey ?? key.split("."),
			queryFn: () => client.call("query", { name: key, payload: params }),
			...opts,
		});

	return {
		data,
		error,
		refetch,

		isSuccess: status === "success",
		isError: status === "error",
		isLoading: !options?.placeholderData && status === "pending",
		isFetched,
		isPlaceholderData,
	};
}
