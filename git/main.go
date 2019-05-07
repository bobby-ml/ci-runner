package git

import (
	"fmt"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"os"
)

func Clone(target string, repo string,  commit string ) (err error){
	os.RemoveAll(target)
	os.MkdirAll(target,0777)
	r, err := git.PlainClone(target, false, &git.CloneOptions{
		URL:     repo,
		Progress: os.Stdout,
	})

	if(err != nil ){
		fmt.Println(err)
		return
	}
	//ref, err := r.Head()
	w, err := r.Worktree()
	err = w.Checkout(&git.CheckoutOptions{
		Hash: plumbing.NewHash(commit),
	})
	return
}