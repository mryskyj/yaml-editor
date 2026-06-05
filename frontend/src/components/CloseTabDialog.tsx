type CloseTabDialogProps = {
	fileName: string;
	message: string;
	onCancel: () => void;
	onConfirm: () => void;
};

export function CloseTabDialog({
	fileName,
	message,
	onCancel,
	onConfirm,
}: CloseTabDialogProps) {
	return (
		<div aria-labelledby="close-tab-dialog-title" className="dialog-backdrop" role="presentation">
			<div aria-modal="true" className="dialog" role="dialog">
				<h2 id="close-tab-dialog-title">未保存のタブを閉じますか？</h2>
				<p className="dialog-file">{fileName}</p>
				<p className="dialog-message">{message}</p>
				<div className="dialog-actions">
					<button className="dialog-button secondary" onClick={onCancel} type="button">
						キャンセル
					</button>
					<button className="dialog-button danger" onClick={onConfirm} type="button">
						保存せずに閉じる
					</button>
				</div>
			</div>
		</div>
	);
}
