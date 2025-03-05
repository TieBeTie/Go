//go:build !solution

package hogwarts

func dfs(course string, courseGraph map[string][]string, visited map[string]rune, courseList []string, time *int) bool {
	visited[course] = 'g'
	for _, prereq := range courseGraph[course] {
		if visited[prereq] == 'g' || visited[prereq] == 'w' && !dfs(prereq, courseGraph, visited, courseList, time) {
			return false
		}
	}
	courseList[*time] = course
	*time++
	visited[course] = 'b'
	return true
}

func GetCourseList(prereqs map[string][]string) []string {
	courseGraph := map[string][]string{}
	for course, prereqs := range prereqs {
		courseGraph[course] = append(courseGraph[course], prereqs...)
		for _, prereq := range prereqs {
			if _, exists := courseGraph[prereq]; !exists {
				courseGraph[prereq] = nil
			}
		}
	}

	visited := map[string]rune{}
	for course := range courseGraph {
		visited[course] = 'w'
	}
	courseList := make([]string, len(courseGraph))
	time := 0

	for course := range courseGraph {
		if visited[course] == 'w' {
			if !dfs(course, courseGraph, visited, courseList, &time) {
				panic("Cyclic dependency detected between courses")
			}
		}
	}
	return courseList
}
