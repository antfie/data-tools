package main

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

const fileAbsolutePathCTEQuery = `
(
	WITH RECURSIVE path_cte(parent_path_id, level, name) AS
	(
		SELECT	f1.path_id,
				f1.level,
				f1.name
		FROM 	files f1
		WHERE 	f1.deleted_at IS NULL
		AND 	f1.ignored = 0
		AND 	id = f.id
		
		UNION
		
		SELECT	p1.parent_path_id,
				p1.level,
				p1.name
		FROM 	files f2
		JOIN 	paths p1 ON f2.path_id = p1.id
		WHERE 	f2.deleted_at IS NULL
		AND 	f2.ignored = 0
		AND		p1.deleted_at IS NULL
		AND 	p1.ignored = 0
		AND 	f2.id = f.id
		
		UNION
		
		SELECT 	p2.parent_path_id,
				p2.level,
				p2.name
		FROM 	paths p2
		JOIN	path_cte ON p2.id = path_cte.parent_path_id
		WHERE 	p2.deleted_at IS NULL
		AND		p2.ignored = 0
	),
	path_ordered AS (
		SELECT 	 	*
		FROM 		path_cte
		ORDER BY	level
	)
	SELECT	group_concat(name, '/')
	FROM	path_ordered
) absolute_path
`

func QueryUnHashedFilePathsWithLimit() string {
	return fmt.Sprintf(`
SELECT		f.id file_id,
			%s
FROM		files f
WHERE 		f.file_hash_id IS NULL
AND 		absolute_path IS NOT NULL
AND 		f.deleted_at IS NULL
AND			f.ignored = 0
ORDER BY	f.id -- for deterministic result order
LIMIT 		?
`, fileAbsolutePathCTEQuery)
}

func QueryGetFileHashesToZapWithLimit() string {
	return fmt.Sprintf(`
SELECT		fh.id file_hash_id,
        	fh.hash,
    		f.id file_id,
			%s
FROM 		file_hashes fh
JOIN  		(
				SELECT		f.id,
							max(f.file_hash_id) AS file_hash_id -- for deterministic result order
				FROM 		files f
				WHERE		f.zapped = 0
				AND			f.deleted_at IS NULL
				AND			f.ignored = 0
				GROUP BY 	f.file_hash_id
  			) f ON f.file_hash_id = fh.id
WHERE		fh.zapped = 0
AND			fh.ignored = 0
ORDER BY	fh.id -- for deterministic result order
LIMIT 		?
`, fileAbsolutePathCTEQuery)
}

func QueryGetDuplicateFilesToZapWithLimit() string {
	return fmt.Sprintf(`
SELECT		f.id file_id,
			%s
FROM 		files f
JOIN 		file_hashes fh ON f.file_hash_id = fh.id
WHERE		f.zapped = 0
AND			f.deleted_at IS NULL
AND			f.ignored = 0
AND			fh.zapped = 1
AND			fh.ignored = 0
ORDER BY	f.size DESC -- to remove the largest duplicates first, and for deterministic result order 
LIMIT 		?
`, fileAbsolutePathCTEQuery)
}

func QueryGetZappedFolders() string {
	return fmt.Sprintf(`
SELECT		%s
FROM 		files f
JOIN 		file_hashes fh ON f.file_hash_id = fh.id
WHERE		f.zapped = 1
AND			f.deleted_at IS NULL
AND			f.ignored = 0
AND			fh.zapped = 1
AND			fh.ignored = 0
ORDER BY	f.id -- for deterministic result order
`, fileAbsolutePathCTEQuery)
}

func QueryGetZappedFileHashesToUnZapWithLimit() string {
	return fmt.Sprintf(`
SELECT		fh.id file_hash_id,
        	fh.hash,
    		f.id file_id,
			%s
FROM 		files f
JOIN 		file_hashes fh ON f.file_hash_id = fh.id
WHERE		f.zapped = 1
AND			f.deleted_at IS NULL
AND			f.ignored = 0
AND			fh.zapped = 1
AND			fh.ignored = 0
AND			f.id NOT IN ?
ORDER BY	f.id -- for deterministic result order
LIMIT 		?
`, fileAbsolutePathCTEQuery)
}

func QueryGetExistingHashSignatures() string {
	return `
SELECT		fh.id hash_id,
    		fh.hash,
        	fh.size,
        	fh.file_type_id,
        	ft.type file_type
FROM		file_hashes fh
JOIN        file_types ft ON fh.file_type_id = ft.id
WHERE 		fh.hash IS NOT NULL
AND 		fh.size IS NOT NULL
AND 		fh.file_type_id IS NOT NULL
ORDER BY	fh.id -- for deterministic result order
`
}

func QueryGetExistingFileTypes() string {
	return `
SELECT		id,
        	type
FROM		file_types
ORDER BY	id -- for deterministic result order
`
}

func (ctx *Context) GetBatchesOfIDs(query string) (int, [][]int, error) {
	total := 0
	var output [][]int
	var batches []string

	if !strings.Contains(query, "BATCH_NUMBER") {
		return 0, nil, errors.New("missing BATCH_NUMBER placeholder in query")
	}

	formattedBatchSize := fmt.Sprintf("(ROW_NUMBER() OVER (ORDER BY id) - 1) / %d AS batch_number", ctx.Config.BatchSize)
	formattedQuery := strings.Replace(query, "BATCH_NUMBER", formattedBatchSize, 1)

	result := ctx.DB.Raw(fmt.Sprintf(`
WITH NumberedRows AS (
	%s
)
SELECT
    GROUP_CONCAT(id) AS ids
FROM NumberedRows
GROUP BY batch_number
ORDER BY batch_number
`, formattedQuery)).Scan(&batches)

	if result.Error != nil {
		return 0, nil, result.Error
	}

	for _, batch := range batches {
		var ids []int

		for _, id := range strings.Split(batch, ",") {
			total++
			idAsInt, err := strconv.Atoi(id)

			if err != nil {
				return 0, nil, err
			}

			ids = append(ids, idAsInt)
		}

		output = append(output, ids)
	}

	return total, output, nil
}
