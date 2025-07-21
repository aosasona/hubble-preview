import Client, {
	type PayloadOf,
	type ResultOf,
	type ProcedureType,
	type Schema,
	type SchemaBasedOnType,
} from "./bindings";
import { ServerAddr } from "./request";

export type ProcedureKey<PType extends ProcedureType> = keyof SchemaBasedOnType<
	Schema,
	PType
>;

export type MutationPayload<T extends ProcedureKey<"mutation">> = PayloadOf<
	Schema,
	"mutation",
	T
>;

export type MutationResult<T extends ProcedureKey<"mutation">> = ResultOf<
	Schema,
	"mutation",
	T
>;

export type QueryPayload<T extends ProcedureKey<"query">> = PayloadOf<
	Schema,
	"query",
	T
>;

export type QueryResult<T extends ProcedureKey<"query">> = ResultOf<
	Schema,
	"query",
	T
>;

const client = Client.new({
	endpoint: `${ServerAddr}/_`,
	fetchOpts: { credentials: "include" },
});

export default client;
