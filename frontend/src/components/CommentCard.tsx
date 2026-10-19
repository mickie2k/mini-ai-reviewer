import { type Comment } from '../type/ReviewResponse';
import CodeMirror, { EditorState, EditorView, lineNumbers } from '@uiw/react-codemirror';
import { vscodeDark } from '@uiw/codemirror-theme-vscode';
import type { Extension } from "@codemirror/state";

export function CommentCard({ languageExtension, comment }: { languageExtension: Extension; comment: Comment }) {

    function sameLine(startLine: number, endLine: number): boolean {
        return startLine === endLine;
    }

    return (
        <div className="border border-neutral-700/70 bg-neutral-900 flex flex-col gap-2 rounded-md p-4">
            <div className="flex justify-between items-center">
                <span className='text-xs inline-flex items-center'>
                    <svg
                        className=
                        "ui-anim-btn-svg mr-2 h-6 w-6 grow-0 fill-neutral-900 duration-400 bg-linear-to-br from-purple-600 to-blue-500 rounded-full p-1"

                        xmlns="http://www.w3.org/2000/svg"
                        viewBox="0 0 24 24"
                        aria-hidden="true"
                    >
                        <path
                            strokeLinecap="round"
                            strokeLinejoin="round"
                            d="M9.813 15.904 9 18.75l-.813-2.846a4.5 4.5 0 0 0-3.09-3.09L2.25 12l2.846-.813a4.5 4.5 0 0 0 3.09-3.09L9 5.25l.813 2.846a4.5 4.5 0 0 0 3.09 3.09L15.75 12l-2.846.813a4.5 4.5 0 0 0-3.09 3.09ZM18.259 8.715 18 9.75l-.259-1.035a3.375 3.375 0 0 0-2.455-2.456L14.25 6l1.036-.259a3.375 3.375 0 0 0 2.455-2.456L18 2.25l.259 1.035a3.375 3.375 0 0 0 2.456 2.456L21.75 6l-1.035.259a3.375 3.375 0 0 0-2.456 2.456ZM16.894 20.567 16.5 21.75l-.394-1.183a2.25 2.25 0 0 0-1.423-1.423L13.5 18.75l1.183-.394a2.25 2.25 0 0 0 1.423-1.423l.394-1.183.394 1.183a2.25 2.25 0 0 0 1.423 1.423l1.183.394-1.183.394a2.25 2.25 0 0 0-1.423 1.423Z"
                        ></path>

                    </svg>
                    AI Reviewer</span>
                <p className="text-right"><span className="font-bold">Line:</span> {sameLine(comment.startLine, comment.endLine) ? comment.startLine : `${comment.startLine} to ${comment.endLine}`}</p>
            </div>

            <p className="text-sm text-neutral-100">{comment.explanation}</p>
            <div className="mt-2 bg-neutral-800 flex flex-col rounded overflow-hidden">
                <span className="text-xs text-neutral-100 p-2">Suggestion Fix</span>
                {comment.fix === "" ? (
                    <div className="text-sm text-neutral-400 p-2">No suggestion provided.</div>
                ) : (
                    <CodeMirror value={comment.fix} className='border-neutral-800' basicSetup={{ lineNumbers: false, highlightActiveLineGutter: false, highlightActiveLine: false }} extensions={[languageExtension, lineNumbers({ formatNumber: (n) => String(n + comment.startLine - 1) }), EditorView.lineWrapping, EditorView.editable.of(false), EditorState.readOnly.of(true)]} theme={vscodeDark} />
                )}
            </div>
        </div>
    )
}