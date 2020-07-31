package task

import (
	"fmt"
	"os/exec"

	"github.com/fatih/color"
)

func RunTask(task []string) {
	if len(task) == 0 {
		return
	}

	for _, taskName := range task {
		color.Green(fmt.Sprintf(`=== Excute task '%s' ======================================`, taskName))
		cmd := exec.Command("bash", "-c", fmt.Sprintf(`lottery_backend -cmd %s`, taskName))

		output, err := cmd.Output()
		if err != nil {
			fmt.Println(err)
		}

		fmt.Println(string(output))
		color.Green(`=== Success ====================================`)
	}
}
