package markdown

const prompt = `
Turn the following unorganized text into a well-structured, fully detailed, and highly readable format while retaining EVERY detail, context, and nuance of the original content.

Your task is to refine the text to improve clarity, grammar, and coherence WITHOUT cutting, summarizing, rephrasing excessively, or omitting ANY information — no matter how minor. Every single detail must be included and accurately represented.

The goal is to enhance readability while ensuring the content remains comprehensive, precise, and faithful to the original by:
	•Organizing the content into logical sections with appropriate subheadings.
	•Using structured formatting (e.g., bullet points, numbered lists) where applicable to present facts, statistics, or comparisons clearly.
	•Emphasizing key terms, names, or headings with bold text to improve navigation and understanding.
	•Maintaining the original tone, humor, and narrative style while ensuring the text remains structured and coherent.
	•Adding clear topic separators or headings to mark distinct sections and transitions, and appropriate code blocks where necessary.

 DO NOT OMIT OR ALTER ANY DETAIL – every fact, statement, or nuance must be preserved exactly as originally intended. Rearrange only to improve clarity and flow, but NEVER at the cost of completeness.

The final output must be generated entirely in [Language] with no use of any other language at any point. The unorganized text itself should not be included in the response.

Text:
`
