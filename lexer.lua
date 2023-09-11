local reporter = require 'reporter'
local Position = require 'position'
local Token = require 'token'
local Type = Token.Type

local lexer = {}

local oneCharToks = {
	['+'] = Type.PLUS,
	['-'] = Type.MINUS,
	['*'] = Type.STAR,
	['/'] = Type.SLASH,
}

---@param source string
function lexer:init(source)
	self.source = source
	self.pos = Position()
	self.curChar = self.source:sub(1,1)
	self.tokens = {}
end

---@return string? oldChar
function lexer:advance()
	local oldChar = self.curChar
	self.pos:advance(oldChar)
	if self.pos.char <= #self.source then
		self.curChar = self.source:sub(self.pos.char, self.pos.char)
	else
		self.curChar = nil
	end

	return oldChar
end

---@return Token
function lexer:number()
	local num = ""
	local start = self.pos:copy()

	while self.curChar and self.curChar:match "%d" do
		num = num .. self:advance()
	end

	return Token(Type.NUMBER, tonumber(num), start, self.pos:copy())
end

---@return Token[]
function lexer:scan()
	while self.curChar do
		while self.curChar and self.curChar:match "%s" do
			self:advance()
		end

		if not self.curChar then break end

		if self.curChar:match "%d" then
			table.insert(self.tokens, self:number())
		elseif oneCharToks[self.curChar] then
			local type = oneCharToks[self.curChar]
			local here = self.pos:copy()
			table.insert(self.tokens, Token(type, self.curChar, here, here))
			self:advance()
		else
			local here = self.pos:copy()
			reporter:error("unknown character '"..self.curChar.."'", here, here)
			self:advance()
		end
	end

	local here = self.pos:copy()
	table.insert(self.tokens, Token(Type.EOF, nil, here, here))
	return self.tokens
end

return lexer
