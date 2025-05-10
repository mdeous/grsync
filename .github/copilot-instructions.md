* Comments should be as concise as possible with a neutral tone;
* When writing documentation, ensure that the writing style and tone is consistent across all the documentation;
* When performing multi-steps operations, always ask for confirmation before proceeding to the next step when intermediate actions are taken between steps;
* Always group similar parts together, for example:
    * In all languages: group globals together, functions together, etc.;
    * In Go: group similarly scoped components together, e.g.: public functions together, private functions together, etc.;
* When making changes, always consider the impact on existing code and ensure that the changes are well integrated into the existing codebase and that all required modifications are made to make them work;
* When making changes that supersede existing code, be careful to not leave any remnants of the old code that could cause confusion or errors;
* When introducing a new dependency, ensure that it is necessary and that the library is well maintained, if no sufficiently well maintained library is available, offer the user different options and let them decide what to do;
* When using a library, always follow the library's documentation and best practices, be creative only if it's strictly required;
* Only add comments when it's useful and helps understanding what a code snippet does, never add comments about past changes (e.g. "moved x to y", "removed z", etc.);
