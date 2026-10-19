import { useState } from 'react'
import './App.css'
import CodeMirror from '@uiw/react-codemirror';
import { LANGUAGE_OPTIONS } from './constants/language';
import { vscodeDark } from '@uiw/codemirror-theme-vscode';
import { type ReviewResponse } from './type/ReviewResponse';
import { type ReviewRequest } from './type/ReviewRequest';
import { ReviewButton } from './components/ReviewButton';
import { LoadingButton } from './components/LoadingButton';
import { CommentCard } from './components/CommentCard';
import axios from 'axios';
import { apiUrl } from './constants/apiUrl';
import { ErrorCard } from './components/ErrorCard';

function App() {
  const [code, setCode] = useState("");
  const [language, setLanguage] = useState("golang");
  const [review, setReviews] = useState<ReviewResponse | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);


  const getErrorMessage = (err: unknown): string => {
    if (axios.isAxiosError(err)) {
      return err.response?.data?.error || err.message;
    }
    return 'An unexpected error occurred.';
  };


  function onClick() {

    const requestBody: ReviewRequest = {
      code: code,
      language: language
    };

    setReviews(null);

    fetchAIReview(requestBody);
  }


  async function fetchAIReview(requestBody: ReviewRequest) {
    try {
      setIsLoading(true);
      setError(null);
      const response = await axios.post(`${apiUrl}/lint`, requestBody);
      setReviews(response.data);
    } catch (error) {
      setError(getErrorMessage(error));
    } finally {
      setIsLoading(false);
    }
  }

  return (
    <div className="w-full max-h-screen h-full overflow-hidden" >
      <div className="grid grid-cols-3 overflow-hidden">
        <section className="col-span-2 border-r overflow-hidden border-neutral-700 text-sm">
          <div className="p-1  h-9 text-xs flex justify-between items-center gap-2 border-b border-neutral-700">
            <span className="text-neutral-400"> &lt;/&gt; Code</span>
            <div className="flex items-center gap-2">
              <span className="text-neutral-400">Language:</span>
              <select id="language" className="block px-2 py-1 rounded-lg placeholder:text-body bg-neutral-700 border-r-6 border-neutral-700 " value={language} onChange={(e) => setLanguage(e.target.value)}>
                {Object.entries(LANGUAGE_OPTIONS).map(([value, { label }]) => (
                  <option key={value} value={value}>{label}</option>
                ))}
              </select>
            </div>
          </div>
          <CodeMirror value={code} placeholder="// Paste your code here..." height="95vh" extensions={[LANGUAGE_OPTIONS[language].extension]} theme={vscodeDark} onChange={(e) => setCode(e)} />
        </section>



        <section className="col-span-1 border-neutral-700 max-h-screen relative">

          <div className="overflow-hidden text-xs">
            <div className="text-sm text-neutral-200 font-bold px-4 border-b border-neutral-700 py-2 h-9">AI Code Review Comments</div>
            <div className="px-4 py-2 overflow-y-scroll max-h-lvh  pb-32">
              <ul className="space-y-2">
                {review && review.comments.map((comment, index) => (
                  <li key={index}>
                    <CommentCard languageExtension={LANGUAGE_OPTIONS[language].extension} comment={comment} />
                  </li>
                ))}

                {review && review.comments.length === 0 && (
                  <li>
                    <CommentCard languageExtension={LANGUAGE_OPTIONS[language].extension} comment={{
                      startLine: 0,
                      endLine: 0,
                      explanation: "No Logic errors found.",
                      fix: ""
                    }} />
                  </li>
                )}

                {error && (
                  <li>
                    <ErrorCard errorMsg={error} />
                  </li>
                )}


                {isLoading && (
                  <span className="relative inline-block animate-pulse text-sm text-neutral-400">
                    AI is reviewing...
                  </span>
                )}
              </ul>

            </div>
          </div>
          <div className="absolute bottom-0 py-2 px-4 border-t border-neutral-700 w-full bg-neutral-800 flex justify-center items-center">
            {!isLoading ? (
              <ReviewButton onClick={onClick} disabled={code.trim() === ""} />
            ) : (
              <LoadingButton title="Reviewing..." />
            )}
          </div>
        </section>
      </div >
    </div >
  )
}



export default App
