import type { EditorDiagnostic } from "../components/ErrorList";

export type EditorCursor = {
	line: number;
	column: number;
};

export type DocumentTab = {
	id: string;
	path: string;
	name: string;
	content: string;
	savedContent: string;
	cursor: EditorCursor;
	diagnostics: EditorDiagnostic[];
};

export type TabState = {
	tabs: DocumentTab[];
	activeTabID: string;
	nextUntitledIndex: number;
};

export type OpenDocumentInput = {
	path?: string;
	name?: string;
	content: string;
};

const emptyCursor = { line: 1, column: 1 };

export function createInitialTabState(content = ""): TabState {
	return {
		tabs: [createUntitledTab("tab-1", 1, content)],
		activeTabID: "tab-1",
		nextUntitledIndex: 2,
	};
}

export function activeTab(state: TabState): DocumentTab | undefined {
	return state.tabs.find((tab) => tab.id === state.activeTabID) ?? state.tabs[0];
}

export function hasUnsavedChanges(tab: DocumentTab): boolean {
	return tab.content !== tab.savedContent;
}

export function isUnsavedTab(tab: DocumentTab): boolean {
	return tab.path === "" || hasUnsavedChanges(tab);
}

export function closeConfirmationMessage(tab: DocumentTab): string {
	if (tab.path === "") {
		return `${tab.name} はまだ保存されていません。保存せずに閉じますか？`;
	}
	return `${tab.name} に未保存の変更があります。保存せずに閉じますか？`;
}

export function addUntitledTab(state: TabState, content = ""): TabState {
	const nextTab = createUntitledTab(
		`tab-${Date.now()}-${state.nextUntitledIndex}`,
		state.nextUntitledIndex,
		content,
	);

	return {
		tabs: [...state.tabs, nextTab],
		activeTabID: nextTab.id,
		nextUntitledIndex: state.nextUntitledIndex + 1,
	};
}

export function fillActiveUntouchedUntitledTab(state: TabState, content: string): TabState {
	const tab = activeTab(state);
	if (!tab || tab.path !== "" || tab.content !== "" || tab.savedContent !== "") {
		return state;
	}
	return updateTab(state, tab.id, (current) => ({
		...current,
		content,
		savedContent: content,
	}));
}

export function openDocumentTab(state: TabState, input: OpenDocumentInput): TabState {
	const path = input.path ?? "";
	if (path) {
		const existing = state.tabs.find((tab) => tab.path === path);
		if (existing) {
			return { ...state, activeTabID: existing.id };
		}
	}

	const tabID = `tab-${Date.now()}-${state.nextUntitledIndex}`;
	const name = input.name ?? fileNameFromPath(path) ?? `untitled-${state.nextUntitledIndex}.yaml`;
	const tab: DocumentTab = {
		id: tabID,
		path,
		name,
		content: input.content,
		savedContent: input.content,
		cursor: emptyCursor,
		diagnostics: [],
	};

	return {
		tabs: [...state.tabs, tab],
		activeTabID: tabID,
		nextUntitledIndex: state.nextUntitledIndex + 1,
	};
}

export function updateActiveContent(state: TabState, content: string): TabState {
	return updateTab(state, state.activeTabID, (tab) => ({ ...tab, content }));
}

export function updateTabCursor(state: TabState, tabID: string, cursor: EditorCursor): TabState {
	return updateTab(state, tabID, (tab) => ({ ...tab, cursor }));
}

export function updateTabDiagnostics(
	state: TabState,
	tabID: string,
	diagnostics: EditorDiagnostic[],
): TabState {
	return updateTab(state, tabID, (tab) => ({ ...tab, diagnostics }));
}

export function markActiveTabSaved(state: TabState, path: string): TabState {
	return updateTab(state, state.activeTabID, (tab) => ({
		...tab,
		path,
		name: fileNameFromPath(path) ?? tab.name,
		savedContent: tab.content,
	}));
}

export function switchTab(state: TabState, tabID: string): TabState {
	if (!state.tabs.some((tab) => tab.id === tabID)) {
		return state;
	}
	return { ...state, activeTabID: tabID };
}

export function switchToAdjacentTab(state: TabState, direction: 1 | -1): TabState {
	const activeIndex = state.tabs.findIndex((tab) => tab.id === state.activeTabID);
	if (activeIndex < 0 || state.tabs.length < 2) {
		return state;
	}

	const nextIndex = (activeIndex + direction + state.tabs.length) % state.tabs.length;
	return {
		...state,
		activeTabID: state.tabs[nextIndex].id,
	};
}

export function closeTab(state: TabState, tabID: string): TabState {
	const tabIndex = state.tabs.findIndex((tab) => tab.id === tabID);
	if (tabIndex < 0) {
		return state;
	}

	const tabs = state.tabs.filter((tab) => tab.id !== tabID);
	if (tabs.length === 0) {
		return {
			...state,
			tabs: [],
			activeTabID: "",
		};
	}
	if (state.activeTabID !== tabID) {
		return { ...state, tabs };
	}

	const nextActive = tabs[Math.min(tabIndex, tabs.length - 1)];
	return {
		...state,
		tabs,
		activeTabID: nextActive.id,
	};
}

export function fileNameFromPath(path: string): string | undefined {
	const normalized = path.replace(/\\/g, "/");
	const parts = normalized.split("/").filter(Boolean);
	return parts.at(-1);
}

function createUntitledTab(id: string, index: number, content: string): DocumentTab {
	const name = index === 1 ? "untitled.yaml" : `untitled-${index}.yaml`;
	return {
		id,
		path: "",
		name,
		content,
		savedContent: content,
		cursor: emptyCursor,
		diagnostics: [],
	};
}

function updateTab(state: TabState, tabID: string, update: (tab: DocumentTab) => DocumentTab): TabState {
	return {
		...state,
		tabs: state.tabs.map((tab) => (tab.id === tabID ? update(tab) : tab)),
	};
}
