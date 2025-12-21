# what it does

when the bot starts it unregisters all previous /commands
from the telegram and registers new ones, so that user can
easily choose the supported commands in UI

A new message comes in from the telegram user to the bot.
bot checks if the user is in whitelist, if not, it does nothing.
next, bot starts processing the message asynchronously and while
the message is being processed, the bot sends a typing indicator
to the user. in case any error occurs, the bot sends that error
to the user (in full if user is admin, otherwise a simplified
error message). admin is simply the 0th user in the whitelist.

## processing the message

the goal of the processing is to append new message to this user's
chat history and pass that history to the llm for processing.
once we get the final response (after any|all tool calls) from
the llm, we append it to the chat history and forward it to the user.

depending on the message type, we first check if our current model
supports that kind of input modality. if not, we first try to convert
the message to a supported modality. if that fails, we throw an error
to be picked up by the error handler and sent to the user. the conversion
is subject to edge cases, such as converting a voice message to text or
a pdf to text or image.

the chat history is preserved in the database per user. the history is serialized
and must be valid across bot restarts and so on. so if the user sends us
a file, we first download it from the telegram server and then convert it
to a supported modality and save it to the database. so that we don't
redownload it and that we can preserve its contents across bot restarts.

## chat history

when we send chat history to llm, we must keep in mind that llms have
limited context length. so we must truncate the history to fit within
the context length. there are several ways to do this:

1. throw out earliest messages until the history fits within the context
window.
2. throw out messages in the middle. we can do this by splitting the
history in half and throwing out messages from the middle until the
history fits.
3. summarize the culled messages and replace them with a single message
that summarizes the content of the culled messages. do this recursively
until the history fits within the context window. culling can be done by
either strategy 1 or 2.
4. use a long term memory to keep track of the important or unique points.
this memory buffer can be appended to the system message to
provide context for the llm. and in remainder, we can send last n messages
to the llm that are still relevant to the current conversation.
