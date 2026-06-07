package dto

type ListSchedulesInput struct {
	Authorization string `header:"Authorization" json:"-"`
	ID            string `path:"id"`
}

type ListSchedulesOutput struct {
	Body struct {
		Schedules []DoseScheduleDTO `json:"schedules"`
	}
}

type CreateScheduleInput struct {
	Authorization string `header:"Authorization" json:"-"`
	ID            string `path:"id"`
	Body          struct {
		Name string `json:"name" minLength:"1" maxLength:"100"`
		Time string `json:"time"`
	}
}

type CreateScheduleOutput struct {
	Body struct {
		Schedule DoseScheduleDTO `json:"schedule"`
	}
}

type GetScheduleInput struct {
	Authorization string `header:"Authorization" json:"-"`
	ID            string `path:"id"`
	ScheduleID    string `path:"scheduleId"`
}

type GetScheduleOutput struct {
	Body struct {
		Schedule DoseScheduleDTO `json:"schedule"`
	}
}

type UpdateScheduleInput struct {
	Authorization string `header:"Authorization" json:"-"`
	ID            string `path:"id"`
	ScheduleID    string `path:"scheduleId"`
	Body          struct {
		Name *string `json:"name,omitempty" minLength:"1" maxLength:"100"`
		Time *string `json:"time,omitempty"`
	}
}

type UpdateScheduleOutput struct {
	Body struct {
		Schedule DoseScheduleDTO `json:"schedule"`
	}
}

type DeleteScheduleInput struct {
	Authorization string `header:"Authorization" json:"-"`
	ID            string `path:"id"`
	ScheduleID    string `path:"scheduleId"`
}

type DeleteScheduleOutput struct {
	Body struct {
		Message string `json:"message"`
	}
}
