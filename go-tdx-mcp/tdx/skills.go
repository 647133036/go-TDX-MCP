package tdx

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

type SkillInfo struct {
	ID          string
	Name        string
	Description string
	Markdown    string
}

// NewSkillPrompt creates an MCP Prompt from a SkillInfo.
func NewSkillPrompt(id, name, description string) mcp.Prompt {
	return mcp.NewPrompt(id,
		mcp.WithPromptTitle(name),
		mcp.WithPromptDescription(description),
	)
}

// AllServerPrompts returns all 45 skills as ServerPrompt entries.
func AllServerPrompts() []server.ServerPrompt {
	skills := AllSkills()
	prompts := make([]server.ServerPrompt, 0, len(skills))

	for _, skill := range skills {
		p := NewSkillPrompt(skill.ID, skill.Name, skill.Description)
		prompts = append(prompts, server.ServerPrompt{
			Prompt:  p,
			Handler: HandleSkillPrompt,
		})
	}
	return prompts
}

// HandleSkillPrompt returns the full markdown skill content for a prompt request.
func HandleSkillPrompt(ctx context.Context, req mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	name := req.Params.Name
	skills := AllSkills()

	for _, skill := range skills {
		if skill.ID == name {
			return &mcp.GetPromptResult{
				Description: skill.Name,
				Messages: []mcp.PromptMessage{
					{
						Role: mcp.RoleUser,
						Content: mcp.NewTextContent(
							fmt.Sprintf("# %s\n\n%s", skill.Name, skill.Markdown),
						),
					},
				},
			}, nil
		}
	}

	return &mcp.GetPromptResult{
		Description: "未找到技能",
		Messages: []mcp.PromptMessage{
			{
				Role: mcp.RoleUser,
				Content: mcp.NewTextContent(
					fmt.Sprintf("技能 '%s' 不存在。请使用 prompts/list 查看可用技能。", name),
				),
			},
		},
	}, nil
}

// AllSkills returns the complete list of 45 investment analysis skills.
func AllSkills() []SkillInfo {
	return []SkillInfo{
		SkillAGZXSB(),
		SkillBKBJ(),
		SkillBoardCPBD(),
		SkillBoardValuation(),
		SkillBXZJXW(),
		SkillCHLTZ(),
		SkillCompanyInfo(),
		SkillCZZDXFXJS(),
		SkillDividendFinancing(),
		SkillDragonTiger(),
		SkillEarningsWarning(),
		SkillEventDrivenShortTerm(),
		SkillFHGDHB(),
		SkillFinancials(),
		SkillFSXYPMBS(),
		SkillGGTZLJYJ(),
		SkillGGWDZK(),
		SkillGGYCBFX(),
		SkillGSZDDF(),
		SkillHotTopic(),
		SkillIndustryChain(),
		SkillIndustryChainMapping(),
		SkillJGCCGDFX(),
		SkillJJZCYJD(),
		SkillLHBXWFG(),
		SkillMainPosition(),
		SkillMRTYJB(),
		SkillPositionDecision(),
		SkillQuantLocal(),
		SkillQuantPython(),
		SkillReportRating(),
		SkillShareCapital(),
		SkillShareholderResearch(),
		SkillStockEvents(),
		SkillTCZQCXX(),
		SkillTradePlan(),
		SkillTradingInfo(),
		SkillValuationPricing(),
		SkillWXDA(),
		SkillWXDBK(),
		SkillWXDETF(),
		SkillWXDJJ(),
		SkillYJYGBY(),
		SkillZJFTJYTL(),
		SkillZTLTBY(),
	}
}
