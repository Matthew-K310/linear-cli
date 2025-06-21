# linear-cli

A bit of a project of necessity for me.

I use [Linear](https://linear.app) for almost all of my project planning, and
while I love their UI, I don't like having to switch context between my
terminal and my browser every time I need to make some changes.

The goal of this project is to create a command line app that will let me
create, edit, and view issues, as well as view my projects and teams. This way,
I can do everything from the comfort of my terminal (which, if we're being
honest, most of us live in the terminal anyway 😅).

###

## Getting Started

Clone the repo

    git clone https://github.com/Matthew-K310/linear-cli.git

Inside of the project directory, create a `.env` file and add your Linear API
key

    API_KEY=<your-key-here>

Run `go build` in the project directory

    go build

From there, add the executable to your ~/.local/bin directory on Unix systems,
and ensure the you have this line in your shell config

    export PATH="$HOME/.local/bin:$PATH"

Then restart your shell session.

###

## Usage

The root command is

    linear-cli

### Creation

You can create an issue with

    linear-cli issues create

This will run a process where you input the issue title and description, and then choose a team, assignee, and status (i.e. todo, in progress, backlog)

### Listing

You can also list issues with

    linear-cli issues list

You can pass in flags to filter the search list

- `-t "<team-name>"` will let you filter by team name
- `-p "<project-name>"` will let you filter by project (dependent on team flag)
- `-s "<status>"` will let you filter by issue status
- `-l "<limit>"` will let you limit the amount of issue responses printed
