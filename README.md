# Build Your Own Audiobooks

OpenAI offers so many opportunities to use the power of large models to create amazing results.
This project uses [OpenAI's text-to-speech (TTS) API](https://platform.openai.com/docs/guides/text-to-speech) to convert text into audiobooks.

Adding Temporal to this process ensures a smooth and robust audiobook creation experience.
Temporal's open-source solutions add hassle-free fault mitigation to your projects.
By integrating Temporal with OpenAI, you can take full advantage of OpenAI's many TTS features without worrying about errors.

OpenAI provides [top-quality voices](https://platform.openai.com/docs/guides/text-to-speech/voice-options) and extensive [language support](https://platform.openai.com/docs/guides/text-to-speech/supported-output-formats), allowing you to choose the perfect tone and rhythm for your audio files.
Together, these two technologies help you focus on creating great content.

**_This project is part of a tutorial._** Follow along by reading the [Build Your Own Audiobooks guide](https://learn.temporal.io/tutorials/go/audiobook/) at [Temporal's learn site](https://learn.temporal.io/tutorials).

## Prerequisites

Before starting this project, make sure you have the following:

- An active OpenAI API developer account and your bearer token.
- Familiarity with the Go programming language.
- The Temporal CLI tool and development server installed on your system.
  Refer to [Getting Started with Go and Temporal](https://learn.temporal.io/getting_started/go/dev_environment/) for detailed setup instructions.

## Setup

Follow these steps to begin converting text to audio.

### Run the Development Server

Make sure the Temporal development server is running and using a persistent store.
Interrupted work can be picked up and continued without repeating steps, even if you experience server interruptions:

```sh
temporal server start-dev \
    --db-filename /path/to/your/temporal.db
```

Once running, connect to the [Temporal Web UI](http://localhost:8233/) and verify that the server is working.


### Instantiate Your Bearer Token

Create an environment variable called OPEN_AI_BEARER_TOKEN to configure your OpenAI credentials.
If you set this value using a shell script, make sure to `source` the script so the variable carries over past the script execution.
The environment variable must be set in the same shell where you'll run your application.

### Run the application (Worker)

1. Build the Worker app:

   ```sh
   go build
   ```

2. Start the app running:

   ```sh
   bin/worker
   ```

If the Worker can't fetch a bearer token from the shell environment, it will loudly fail at launch.
This early check prevents you from running jobs and waiting to find out that you forgot to set the bearer token until you're well into the Workflow process.

Now that the Worker is running, you can submit jobs for text-to-speech processing.

## Submit narration jobs

In a new terminal window, use the Temporal CLI tool to build audio from text files.
Use the Workflow `execute` subcommand to watch the execution in real time from the command line:

```
temporal workflow execute \
    --type TTSWorkflow \
    --task-queue tts-task-queue \
    --input '"/path/to/your/text-file.txt"' \
    --workflow-id "tristam-shandy-tts"
```

* **type**: The name of this text-to-speech Workflow is `TTSWorkflow`.
* **task-queue**: This Worker polls the "tts-task-queue" Task Queue.
* **input**: Pass a quoted JSON string with a /path/to/your/input/text-file.
* **workflow-id**: Set a descriptive name for your Workflow Id.
  This makes it easier to track your Workflow Execution in the Web UI.
 
Please note:

* The identifier you set won't affect the input text file or the output audio file names.
* Use full paths as the Worker is not given context for relative path resolution.
* Sample files appear in the `text-samples` folder.

### The output file

Your output is collected in a system-provided temporary file.

After, your generated MP3 audio is moved into the same folder as your input text file.
It uses the same name replacing the `txt` extension with `mp3`.
If an output file already exists, the project versions it to prevent name collisions.

The Workflow returns a string, the /path/to/your/output/audio-file.
Check the Web UI Input and Results section after the Workflow completes.
The results path is also displayed as part of the CLI's `workflow execute` command output and in the Worker logs.

### Cautions and notes

- Do not modify your input or output files while the Workflow is running.
- The Workflow fails if you don't pass a valid text file named with a `txt` extension.

### Peeking at the process

This project includes a Query to check progress during long processes.
Run it in a separate terminal window or tab:

```
temporal workflow query \
    --type fetchMessage \
    --workflow-id YourWorkflowId
```

### Validate your audio output

The open source [checkmate](https://github.com/Sjord/checkmate) app lets you validate your generated MP3 file for errors.

```
$ mpck -v audio.mp3

SUMMARY: audio.mp3
    version                       MPEG v2.0
    layer                         3
    bitrate                       160000 bps
    samplerate                    24000 Hz
    frames                        23723
    time                          9:29.352
    unidentified                  0 b (0%)
    stereo                        yes
    size                          11120 KiB
    ID3V1                         no
    ID3V2                         no
    APEV1                         no
    APEV2                         no
    last frame                    
        offset                    11386560 b (0xadbec0)
        length                    480
    errors                        none
    result                        Ok
```

### Converting chapter files into a book

Consider submitting each chapter as a separate Workflow.
This allows you to quality check each file and ensure it "reads" the way you want.
After generating all your chapters and front material, back material, and other book elements, you can combine them together.

Use [`ffmpeg`](https://ffmpeg.org) to combine audio files.

1. Create a text file listing the files, for example:

```text
file 'title_sequence.mp3'
file 'introduction.mp3'
file 'chapter1.mp3'
file 'chapter2.mp3'
file 'chapter3.mp3'
...
``` 

2. Perform the concatenation:

```
ffmpeg -f concat -safe 0 -i chapters-list.txt -c copy fullbook.mp3
```

3. (optional) `ffmpeg` allows you to convert your audio to other formats.
For example:

```
ffmpeg -i fullbook.mp3 fullbook.m4a
```

## Project Structure

```sh
.
├── LICENSE
├── README.md
├── TTSActivities.go
├── TTSWorkflow.go
├── go.mod
├── go.sum
├── text-samples
│   ├── austen.txt
│   └── doyle.txt
└── worker
    └── main.go
```
