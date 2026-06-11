#!/usr/bin/env bash
### coded for the use of -> alt+shift+r
# this will create a run window if it not exist in this tmux session
# and than restarts the go run main.go

#break below pane if there are more than 1 on window 1
#what should be the run pane into seperate window again
#this is needed to rerun it on a window named RUN
panes=$(tmux list-panes -t :1 | wc -l)
if [ "$panes" -gt 1 ]; then
	tmux break-pane -s 1 -n run
	sleep 0.02
fi

SESSION=$(tmux display-message -p '#S')
WINDOW_NAME="run"
CURRENT_PATH=$(pwd)

# List only window names and check for exact match
if tmux list-windows -t "$SESSION" -F "#{window_name}" | grep -Fxq "$WINDOW_NAME"; then
	# Window exists: select it
	tmux select-window -t "$SESSION:$WINDOW_NAME"
else
	# Window does not exist: create it
	tmux new-window -t "$SESSION" -n "$WINDOW_NAME" -c "$CURRENT_PATH"
fi

#tmux send-keys -t "$SESSION:$WINDOW_NAME" C-c "./go run main.go" C-m
tmux send-keys C-c C-m C-m C-m C-m
sleep 0.02

tmux send-keys "bash -c '
go run .
'" C-m

#tmux send-keys 'sqlc generate -f ./db/sqlc.yaml && migrate -path db/migrations -database "sqlite3://db.sqlite" up && sqlite3 db.sqlite ".schema" > db/schema.sql && sqlc generate -f ./db/sqlc.yaml && go run ./cmd/app/' C-m
#tmux send-keys 'go run ./cmd/app/' C-m

sleep 0.02
tmux last-window

#this will join the rsun window into first window
#if you don't want this commant this out
sleep 0.02
tmux join-pane -h -s :run -t :1 -p 32
#set focus to pane 0 -> what should be vim
tmux select-pane -t 0
