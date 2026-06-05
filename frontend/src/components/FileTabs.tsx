import { isUnsavedTab, type DocumentTab } from "../editor/tabs";

type FileTabsProps = {
	tabs: DocumentTab[];
	activeTabID: string;
	onSelect: (tabID: string) => void;
	onClose: (tabID: string) => void;
};

export function FileTabs({ tabs, activeTabID, onSelect, onClose }: FileTabsProps) {
	return (
		<div className="file-tabs" role="tablist" aria-label="Open YAML files">
			{tabs.map((tab) => {
				const isActive = tab.id === activeTabID;
				const isUnsaved = isUnsavedTab(tab);
				return (
					<div
						className={`file-tab-wrap${isActive ? " active" : ""}${isUnsaved ? " dirty" : ""}`}
						key={tab.id}
						title={tab.path || tab.name}
					>
						<button
							aria-selected={isActive}
							className="file-tab"
							onClick={() => onSelect(tab.id)}
							role="tab"
							type="button"
						>
							<span className="file-tab-name">{tab.name}</span>
							{isUnsaved ? <span aria-label="Unsaved" className="file-tab-dirty" /> : null}
						</button>
						<button
							aria-label={`Close ${tab.name}`}
							className="file-tab-close"
							onClick={() => {
								onClose(tab.id);
							}}
							type="button"
						>
							x
						</button>
					</div>
				);
			})}
		</div>
	);
}
