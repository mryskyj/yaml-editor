import { useEffect, useState } from "react";

type ScheduleMenuProps = {
	template: string;
	onCancel: () => void;
	onReset: () => void;
	onSave: (template: string) => void;
};

export function ScheduleMenu({
	template,
	onCancel,
	onReset,
	onSave,
}: ScheduleMenuProps) {
	const [draft, setDraft] = useState(template);

	useEffect(() => {
		setDraft(template);
	}, [template]);

	return (
		<div className="dialog-backdrop" role="presentation">
			<section className="dialog schedule-dialog" role="dialog" aria-modal="true" aria-labelledby="schedule-menu-title">
				<h2 id="schedule-menu-title">Schedules</h2>
				<textarea
					aria-label="Schedule template"
					className="schedule-textarea"
					value={draft}
					onChange={(event) => setDraft(event.target.value)}
				/>
				<div className="dialog-actions">
					<button type="button" className="dialog-button secondary" onClick={onReset}>
						Reset
					</button>
					<button type="button" className="dialog-button secondary" onClick={onCancel}>
						Cancel
					</button>
					<button type="button" className="dialog-button primary" onClick={() => onSave(draft)}>
						Save
					</button>
				</div>
			</section>
		</div>
	);
}
