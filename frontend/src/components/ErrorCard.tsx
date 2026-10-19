export function ErrorCard({ errorMsg }: { errorMsg: string }) {


    return (
        <div className="border border-neutral-700 bg-neutral-900 flex flex-col gap-2 rounded-md p-4">
            <div className="flex flex-col gap-2">
                <span className='text-sm font-bold text-red-400 inline-flex items-center'>Error</span>
                <p className="text-sm text-neutral-100">{errorMsg}. <br /> Please try again.</p>
            </div>
        </div >
    )
}