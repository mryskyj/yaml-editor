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

export function createInitialTabState(): TabState {
	return {
		tabs: [createUntitledTab("tab-1", 1)],
		activeTabID: "tab-1",
		nextUntitledIndex: 2,
	};
}

export function activeTab(state: TabState): DocumentTab {
	return state.tabs.find((tab) => tab.id === state.activeTabID) ?? state.tabs[0];
}

export function hasUnsavedChanges(tab: DocumentTab): boolean {
	return tab.content !== tab.savedContent;
}

export function addUntitledTab(state: TabState): TabState {
	const nextTab = createUntitledTab(
		`tab-${Date.now()}-${state.nextUntitledIndex}`,
		state.nextUntitledIndex,
	);

	return {
		tabs: [...state.tabs, nextTab],
		activeTabID: nextTab.id,
		nextUntitledIndex: state.nextUntitledIndex + 1,
	};
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

export function closeTab(state: TabState, tabID: string): TabState {
	const tabIndex = state.tabs.findIndex((tab) => tab.id === tabID);
	if (tabIndex < 0) {
		return state;
	}

	const tabs = state.tabs.filter((tab) => tab.id !== tabID);
	if (tabs.length === 0) {
		return createInitialTabState();
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

function createUntitledTab(id: string, index: number): DocumentTab {
	const name = index === 1 ? "untitled.yaml" : `untitled-${index}.yaml`;
	return {
		id,
		path: "",
		name,
		content: "",
		savedContent: "",
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
