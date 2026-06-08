export type SelectContext = {
	value: string;
	setValue: (value: string) => void;
	registerLabel: (value: string, label: string) => void;
	getLabel: (value: string) => string;
};

export const SELECT_CONTEXT_KEY = Symbol('select-context');
