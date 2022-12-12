/*
 * Copyright (c) 2022, Gideon Williams <gideon@gideonw.com>
 *
 * SPDX-License-Identifier: BSD-2-Clause
 */

package proto

type Job struct {
	ID          string `json:"id"`
	URL         string `json:"url"`
	Req         int    `json:"req"`
	Concurrency int    `json:"concurrency"`
	Duration    int    `json:"duration"`
	Rate        int    `json:"rate"`
}
