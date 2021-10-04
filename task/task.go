package task

import (
	"fmt"
	"os/exec"

	"github.com/fatih/color"
)

func RunTask(task []string, taskCmd string) {
	if len(task) == 0 {
		return
	}

	if taskCmd == "" {
		fmt.Println("=== Task config 'taskCommand' undefined . Skip ============")
		return
	}

	for _, taskName := range task {
		color.Green(fmt.Sprintf(`=== Excute task '%s' ======================`, taskName))
		cmd := exec.Command("bash", "-c", fmt.Sprintf(`%s %s`, taskCmd, taskName))

		output, err := cmd.Output()
		if err != nil {
			color.Red("Failed command: " + fmt.Sprintf(`%s %s`, taskCmd, taskName))
			panic(err)
		}

		fmt.Println(string(output))
		color.Green(`=== Success ====================================`)
	}
}
