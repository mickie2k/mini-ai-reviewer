import { javascript } from "@codemirror/lang-javascript";
import { python } from "@codemirror/lang-python";
import { go } from "@codemirror/lang-go";
import { java } from "@codemirror/lang-java";
import type { Extension } from "@codemirror/state";

export interface LanguageConfig {
    label: string;
    extension: Extension;
}

export const LANGUAGE_OPTIONS: Record<string, LanguageConfig> = {
    golang: {
        label: "Go",
        extension: go(),
    },
    python: {
        label: "Python",
        extension: python(),
    },
    java: {
        label: "Java",
        extension: java(),
    },
    javascript: {
        label: "JavaScript",
        extension: javascript({ jsx: true }),
    },
    typescript: {
        label: "TypeScript",
        extension: javascript({ jsx: true, typescript: true }),
    },
};
