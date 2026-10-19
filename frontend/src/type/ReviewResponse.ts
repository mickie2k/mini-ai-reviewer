export type ReviewResponse = {
    comments: Comment[];
};

export type Comment = {
    startLine: number;
    endLine: number;
    explanation: string;
    fix: string;
};
